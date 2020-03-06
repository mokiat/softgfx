package conversion

import (
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/mokiat/go-data-front/decoder/obj"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/softgfx/cmd/softgfx-lvlgen/internal/bsp"
	"github.com/mokiat/softgfx/cmd/softgfx-lvlgen/internal/objutil"
	"github.com/mokiat/softgfx/cmd/softgfx-lvlgen/internal/scene"
	"github.com/mokiat/softgfx/internal/data"
)

const precision = 0.001

func run(in io.Reader, out io.Writer, scale float64) error {
	decoder := obj.NewDecoder(obj.DefaultLimits())
	model, err := decoder.Decode(in)
	if err != nil {
		return fmt.Errorf("failed to decode obj file: %w", err)
	}

	log.Printf("scaling model (factor: %f)...\n", scale)
	objutil.Model(model).Scale(scale)

	log.Println("extracting vertical lines...")
	verticalLines := extractVerticalLines(model)
	log.Printf("\tfound: %d\n", len(verticalLines))
	verticalLines = verticalLines.Dedupe(precision)
	log.Printf("\tunique: %d\n", len(verticalLines))

	log.Println("extracting floor triangles...")
	floorTriangles := extractFloorTriangles(model)
	log.Printf("\tfound: %d\n", len(floorTriangles))

	log.Println("extracting ceiling triangles...")
	ceilingTriangles := extractCeilingTriangles(model)
	log.Printf("\tfound: %d\n", len(ceilingTriangles))

	log.Println("extracting vertical triangles...")
	verticalTriangles := extractVerticalTriangles(model)
	log.Printf("\tfound: %d\n", len(verticalTriangles))

	log.Println("building segments...")
	segments := buildSegments(verticalTriangles)
	log.Printf("\ttotal: %d\n", len(segments))
	segments = partitionSegments(segments, verticalLines)
	log.Printf("\tpartitioned: %d\n", len(segments))

	log.Println("building blocks...")
	blocks := buildBlocks(segments)
	log.Printf("\ttotal: %d\n", len(blocks))
	blocks = blocks.Merge(precision)
	log.Printf("\tmerged: %d\n", len(blocks))

	log.Println("building walls...")
	walls := buildWalls(blocks, floorTriangles, ceilingTriangles)
	log.Printf("\ttotal: %d\n", len(walls))

	log.Println("partitioning walls...")
	tree := bsp.Partition(walls, precision)
	log.Printf("\ttotal: %d\n", tree.Count())

	jsonLevel := buildLevel(tree)
	if err := json.NewEncoder(out).Encode(jsonLevel); err != nil {
		return fmt.Errorf("failed to encode json level: %w", err)
	}
	return nil
}

func extractVerticalLines(model *obj.Model) scene.VerticalLineList {
	var result scene.VerticalLineList
	for objLine := range objutil.Model(model).Edges() {
		if !objLine.Vertical(precision) {
			continue
		}
		result = append(result, scene.VerticalLine{
			X: objLine.P1.X,
			Z: objLine.P1.Z,
		})
	}
	return result
}

func extractFloorTriangles(model *obj.Model) scene.TriangleList {
	var result scene.TriangleList
	for objTriangle := range objutil.Model(model).Triangles() {
		sceneTriangle := scene.Triangle{
			P1:          objTriangle.P1,
			P2:          objTriangle.P2,
			P3:          objTriangle.P3,
			TextureName: objTriangle.MaterialName,
		}
		if sceneTriangle.IsFloor(precision) {
			result = append(result, sceneTriangle)
		}
	}
	return result
}

func extractCeilingTriangles(model *obj.Model) scene.TriangleList {
	var result scene.TriangleList
	for objTriangle := range objutil.Model(model).Triangles() {
		sceneTriangle := scene.Triangle{
			P1:          objTriangle.P1,
			P2:          objTriangle.P2,
			P3:          objTriangle.P3,
			TextureName: objTriangle.MaterialName,
		}
		if sceneTriangle.IsCeiling(precision) {
			result = append(result, sceneTriangle)
		}
	}
	return result
}

func extractVerticalTriangles(model *obj.Model) scene.TriangleList {
	var result scene.TriangleList
	for objTriangle := range objutil.Model(model).Triangles() {
		sceneTriangle := scene.Triangle{
			P1:          objTriangle.P1,
			P2:          objTriangle.P2,
			P3:          objTriangle.P3,
			TextureName: objTriangle.MaterialName,
		}
		if sceneTriangle.IsVertical(precision) {
			result = append(result, sceneTriangle)
		}
	}
	return result
}

func buildSegments(verticalTriangles scene.TriangleList) scene.SegmentList {
	result := make(scene.SegmentList, len(verticalTriangles))
	for i, triangle := range verticalTriangles {
		result[i] = scene.Segment{
			Left:   triangle.Left(),
			Right:  triangle.Right(),
			Normal: triangle.Normal(),
			Lines: []scene.Line{
				triangle.Line1(),
				triangle.Line2(),
				triangle.Line3(),
			},
			TextureName: triangle.TextureName,
		}
	}
	return result
}

func partitionSegments(segments scene.SegmentList, verticalLines scene.VerticalLineList) scene.SegmentList {
	var result scene.SegmentList
	for _, segment := range segments {
		subSegments := partitionSegment(segment, verticalLines)
		result = append(result, subSegments...)
	}
	return result
}

func partitionSegment(segment scene.Segment, verticalLines scene.VerticalLineList) scene.SegmentList {
	for i, verticalLine := range verticalLines {
		isPartitioned :=
			!verticalLine.Equal(segment.Left, precision) &&
				!verticalLine.Equal(segment.Right, precision) &&
				segment.ContainsVerticalLine(verticalLine, precision)

		if isPartitioned {
			leftSegment := segment.LeftPartition(verticalLine, precision)
			leftSubSegments := partitionSegment(leftSegment, verticalLines[i+1:])

			rightSegment := segment.RightPartition(verticalLine, precision)
			rightSubSegments := partitionSegment(rightSegment, verticalLines[i+1:])

			return append(leftSubSegments, rightSubSegments...)
		}
	}

	return scene.SegmentList{segment}
}

func buildBlocks(segments scene.SegmentList) scene.BlockList {
	var result scene.BlockList
	for _, segment := range segments {
		if len(segment.Lines) == 0 {
			log.Println("warning: skipping segment: no lines present")
			continue
		}
		result = append(result, scene.Block{
			Left:   segment.Left,
			Right:  segment.Right,
			Normal: segment.Normal,
			Spans: []scene.Span{
				scene.Span{
					Top:         segment.Top(),
					Bottom:      segment.Bottom(),
					TextureName: segment.TextureName,
				},
			},
		})
	}
	return result
}

func buildWalls(blocks scene.BlockList, floorTriangles, ceilingTriangles scene.TriangleList) []*bsp.Wall {
	var walls []*bsp.Wall
	for _, block := range blocks {
		wall, err := buildWall(block, floorTriangles, ceilingTriangles)
		if err != nil {
			log.Printf("warning: skipping block: %v\n", err)
			continue
		}
		walls = append(walls, wall)
	}
	return walls
}

func buildWall(block scene.Block, floorTriangles, ceilingTriangles scene.TriangleList) (*bsp.Wall, error) {
	wall := &bsp.Wall{
		LeftX:  block.Left.X,
		LeftZ:  block.Left.Z,
		RightX: block.Right.X,
		RightZ: block.Right.Z,
	}

	switch spanCount := len(block.Spans); spanCount {
	case 2:
		// it is a split wall
		outerCeiling, outerCeilingFound := findHorizontalTriangle(ceilingTriangles, spanMiddleTop(block, block.Spans[0]))
		innerCeiling, innerCeilingFound := findHorizontalTriangle(ceilingTriangles, spanMiddleBottom(block, block.Spans[0]))
		innerFloor, innerFloorFound := findHorizontalTriangle(floorTriangles, spanMiddleTop(block, block.Spans[1]))
		outerFloor, outerFloorFound := findHorizontalTriangle(floorTriangles, spanMiddleBottom(block, block.Spans[1]))
		if !outerCeilingFound || !innerCeilingFound || !innerFloorFound || !outerFloorFound {
			return nil, fmt.Errorf("could not find all floors and ceilings for block")
		}
		wall.Ceiling = &bsp.Extrusion{
			Top:              block.Spans[0].Top,
			Bottom:           block.Spans[0].Bottom,
			OuterTextureName: outerCeiling.TextureName,
			FaceTextureName:  block.Spans[0].TextureName,
			InnerTextureName: innerCeiling.TextureName,
		}
		wall.Floor = &bsp.Extrusion{
			Top:              block.Spans[1].Top,
			Bottom:           block.Spans[1].Bottom,
			InnerTextureName: innerFloor.TextureName,
			FaceTextureName:  block.Spans[1].TextureName,
			OuterTextureName: outerFloor.TextureName,
		}
		return wall, nil

	case 1:
		// it is a solid wall or a singular extrusion
		outerCeiling, outerCeilingFound := findHorizontalTriangle(ceilingTriangles, spanMiddleTop(block, block.Spans[0]))
		outerFloor, outerFloorFound := findHorizontalTriangle(floorTriangles, spanMiddleBottom(block, block.Spans[0]))

		switch {
		case outerCeilingFound && outerFloorFound:
			// it is a solid wall top-to-bottom
			wall.Ceiling = &bsp.Extrusion{
				Top:              block.Spans[0].Top,
				Bottom:           (block.Spans[0].Top + block.Spans[0].Bottom) / 2.0,
				OuterTextureName: outerCeiling.TextureName,
				FaceTextureName:  block.Spans[0].TextureName,
				InnerTextureName: outerCeiling.TextureName, // irrelevant, but set to something valid
			}
			wall.Floor = &bsp.Extrusion{
				Top:              (block.Spans[0].Top + block.Spans[0].Bottom) / 2.0,
				Bottom:           block.Spans[0].Bottom,
				InnerTextureName: outerFloor.TextureName, // irrelevant, but set to something valid
				FaceTextureName:  block.Spans[0].TextureName,
				OuterTextureName: outerFloor.TextureName,
			}
			return wall, nil

		case outerCeilingFound:
			// it is a ceiling wall extrusion
			innerCeiling, innerCeilingFound := findHorizontalTriangle(ceilingTriangles, spanMiddleBottom(block, block.Spans[0]))
			if !innerCeilingFound {
				return nil, fmt.Errorf("could not find inner ceiling texture for ceiling extrusion")
			}
			wall.Ceiling = &bsp.Extrusion{
				Top:              block.Spans[0].Top,
				Bottom:           block.Spans[0].Bottom,
				OuterTextureName: outerCeiling.TextureName,
				FaceTextureName:  block.Spans[0].TextureName,
				InnerTextureName: innerCeiling.TextureName,
			}
			return wall, nil

		case outerFloorFound:
			// it is a floor wall extrusion
			innerFloor, innerFloorFound := findHorizontalTriangle(floorTriangles, spanMiddleTop(block, block.Spans[0]))
			if !innerFloorFound {
				return nil, fmt.Errorf("could not find inner floor texture for floor extrusion")
			}
			wall.Floor = &bsp.Extrusion{
				Top:              block.Spans[0].Top,
				Bottom:           block.Spans[0].Bottom,
				InnerTextureName: innerFloor.TextureName,
				FaceTextureName:  block.Spans[0].TextureName,
				OuterTextureName: outerFloor.TextureName,
			}
			return wall, nil

		default:
			return nil, fmt.Errorf("could not find floor or ceiling texture for block")
		}

	default:
		return nil, fmt.Errorf("unexpected span count: %d", spanCount)
	}
}

func spanMiddleTop(block scene.Block, span scene.Span) dprec.Vec3 {
	return dprec.Vec3{
		X: (block.Left.X + block.Right.X) / 2.0,
		Y: span.Top,
		Z: (block.Left.Z + block.Right.Z) / 2.0,
	}
}

func spanMiddleBottom(block scene.Block, span scene.Span) dprec.Vec3 {
	return dprec.Vec3{
		X: (block.Left.X + block.Right.X) / 2.0,
		Y: span.Bottom,
		Z: (block.Left.Z + block.Right.Z) / 2.0,
	}
}

func findHorizontalTriangle(triangles scene.TriangleList, edgePoint dprec.Vec3) (scene.Triangle, bool) {
	for _, triangle := range triangles {
		if triangle.Line1().ContainsPoint(edgePoint, precision) {
			return triangle, true
		}
		if triangle.Line2().ContainsPoint(edgePoint, precision) {
			return triangle, true
		}
		if triangle.Line3().ContainsPoint(edgePoint, precision) {
			return triangle, true
		}
	}
	return scene.Triangle{}, false
}

func buildLevel(root *bsp.Wall) data.Level {
	jsonTextures := make([]string, 0)
	jsonWalls := make([]data.Wall, 0, root.Count())

	registerTexture := func(textureName string) int {
		for i, jsonTexture := range jsonTextures {
			if jsonTexture == textureName {
				return i
			}
		}

		jsonTextures = append(jsonTextures, textureName)
		return len(jsonTextures) - 1
	}

	var processWall func(wall *bsp.Wall) int
	processWall = func(wall *bsp.Wall) int {
		if wall == nil {
			return -1
		}
		index := len(jsonWalls)
		jsonWalls = append(jsonWalls, data.Wall{})
		jsonWall := data.Wall{
			LeftEdgeX:  float32(wall.LeftX),
			LeftEdgeZ:  -float32(wall.LeftZ),
			RightEdgeX: float32(wall.RightX),
			RightEdgeZ: -float32(wall.RightZ),
			FrontWall:  processWall(wall.Front),
			BackWall:   processWall(wall.Back),
		}
		if wall.Floor != nil {
			jsonWall.Floor = &data.Extrusion{
				Top:          -float32(wall.Floor.Top),
				Bottom:       -float32(wall.Floor.Bottom),
				OuterTexture: registerTexture(wall.Floor.OuterTextureName),
				FaceTexture:  registerTexture(wall.Floor.FaceTextureName),
				InnerTexture: registerTexture(wall.Floor.InnerTextureName),
			}
		}
		if wall.Ceiling != nil {
			jsonWall.Ceiling = &data.Extrusion{
				Top:          -float32(wall.Ceiling.Top),
				Bottom:       -float32(wall.Ceiling.Bottom),
				InnerTexture: registerTexture(wall.Ceiling.InnerTextureName),
				FaceTexture:  registerTexture(wall.Ceiling.FaceTextureName),
				OuterTexture: registerTexture(wall.Ceiling.OuterTextureName),
			}
		}
		jsonWalls[index] = jsonWall
		return index
	}
	processWall(root)

	return data.Level{
		Textures: jsonTextures,
		Walls:    jsonWalls,
	}
}
