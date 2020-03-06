package objutil

import (
	"log"

	"github.com/mokiat/go-data-front/decoder/obj"
	"github.com/mokiat/gomath/dprec"
)

type Triangle struct {
	P1           dprec.Vec3
	P2           dprec.Vec3
	P3           dprec.Vec3
	MaterialName string
}

type Line struct {
	P1 dprec.Vec3
	P2 dprec.Vec3
}

func (l Line) Vertical(precision float64) bool {
	return dprec.EqEps(l.P1.X, l.P2.X, precision) && dprec.EqEps(l.P1.Z, l.P2.Z, precision)
}

func Model(model *obj.Model) ModelWrapper {
	return ModelWrapper{
		model: model,
	}
}

type ModelWrapper struct {
	model *obj.Model
}

func (w ModelWrapper) Scale(scale float64) {
	for i := range w.model.Vertices {
		w.model.Vertices[i].X *= scale
		w.model.Vertices[i].Y *= scale
		w.model.Vertices[i].Z *= scale
	}
}

func (w ModelWrapper) Meshes() <-chan *obj.Mesh {
	result := make(chan *obj.Mesh)
	go func() {
		for _, obj := range w.model.Objects {
			for _, mesh := range obj.Meshes {
				result <- mesh
			}
		}
		close(result)
	}()
	return result
}

func (w ModelWrapper) Triangles() <-chan Triangle {
	result := make(chan Triangle)
	go func() {
		for mesh := range w.Meshes() {
			for _, face := range mesh.Faces {
				vertexCount := len(face.References)
				if vertexCount < 3 {
					log.Printf("warning: skipping face: insufficient number of vertices: %d\n", vertexCount)
					continue
				}

				vertex1 := w.model.GetVertexFromReference(face.References[0])
				vertex2 := w.model.GetVertexFromReference(face.References[1])
				for i := 2; i < vertexCount; i++ {
					vertex3 := w.model.GetVertexFromReference(face.References[i])
					result <- Triangle{
						P1:           dprec.NewVec3(vertex1.X, vertex1.Y, vertex1.Z),
						P2:           dprec.NewVec3(vertex2.X, vertex2.Y, vertex2.Z),
						P3:           dprec.NewVec3(vertex3.X, vertex3.Y, vertex3.Z),
						MaterialName: mesh.MaterialName,
					}
					vertex2 = vertex3
				}
			}
		}
		close(result)
	}()
	return result
}

func (w ModelWrapper) Edges() <-chan Line {
	result := make(chan Line)
	go func() {
		for triangle := range w.Triangles() {
			result <- Line{
				P1: triangle.P1,
				P2: triangle.P2,
			}
			result <- Line{
				P1: triangle.P2,
				P2: triangle.P3,
			}
			result <- Line{
				P1: triangle.P3,
				P2: triangle.P1,
			}
		}
		close(result)
	}()
	return result
}
