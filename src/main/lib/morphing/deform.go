package morphing

import (
	"../utils"
	"image"
	"image/draw"
	"math"
)

type MeshDeformation struct {
	initialMesh, finalMesh *Mesh
	primitiveMapping       map[Triangle]Triangle
}

func NewMeshDeformation(mesh *Mesh, dst image.Image, vertexMapping map[Vertex]Vertex, t float64) *MeshDeformation {
	triangles := make([]Triangle, len(mesh.Triangles))
	primitiveMapping := make(map[Triangle]Triangle)
	for idxTri, tri := range mesh.Triangles {
		newPoints := make([]Vertex, 3)

	vertexLoop:
		for idxPoint, p := range tri.points {
			for key, val := range vertexMapping {
				if p.Equal(key) {
					newPoints[idxPoint] = Vx(utils.LERP(p.X, val.X, t), utils.LERP(p.Y, val.Y, t))
					continue vertexLoop
				}
			}

			newPoints[idxPoint] = p
		}

		triangles[idxTri] = *NewTriangle(newPoints[0], newPoints[1], newPoints[2])
		primitiveMapping[triangles[idxTri]] = mesh.Triangles[idxTri]
	}

	finalMesh := NewMesh(dst)
	finalMesh.Triangles = triangles

	mapping := new(MeshDeformation)
	mapping.initialMesh = mesh
	mapping.finalMesh = finalMesh
	mapping.primitiveMapping = primitiveMapping
	return mapping
}

func (mesh *MeshDeformation) Deform() image.Image {
	dstBounds := mesh.finalMesh.Texture.Bounds()
	ret := utils.CreateRGBA(dstBounds)
	dstEdges, _ := mesh.finalMesh.Edges()

	// define helper function
	collide := func(e Edge, y float64) Vertex {
		A, B := e.End.Y-e.Start.Y, e.Start.X-e.End.X
		C := (e.End.X * e.Start.Y) - (e.Start.X * e.End.Y)
		x := -(B*y + C) / A
		return Vx(x, y)
	}

	type CollisionTriangle struct {
		startVertex, endVertex Vertex
		startEdge, endEdge     Edge
		triangle               Triangle
	}

	for y := dstBounds.Min.Y; y < dstBounds.Max.Y; y++ {
		Y := float64(y) + 0.5

		collisionEdges := make([]Edge, 0)
		collisionVertexes := make([]Vertex, 0)

		// find edges that have a collision point with the current line
		for _, e := range dstEdges {
			// horizontal edges should be ignored
			if e.Start.Y == e.End.Y {
				continue
			}

			// if both defining points of an edge are on the same side
			// of the horizontal line, there cannot be any collision
			if !(e.Start.Y >= Y && e.End.Y <= Y) && !(e.Start.Y <= Y && e.End.Y >= Y) {
				continue
			}

			collisionEdges = append(collisionEdges, e)
			collisionVertexes = append(collisionVertexes, collide(e, Y))
		}

		// sort the collision edges and vertexes according to the collision
		// point's X coordinate
		for i := 0; i < len(collisionVertexes); i++ {
			for j := i + 1; j < len(collisionVertexes); j++ {
				if collisionVertexes[i].X > collisionVertexes[j].X {
					collisionEdges[i], collisionEdges[j] = collisionEdges[j], collisionEdges[i]
					collisionVertexes[i], collisionVertexes[j] = collisionVertexes[j], collisionVertexes[i]
				}
			}
		}

		// find all triangles the line collides
		collisionTriangles := make([]CollisionTriangle, len(collisionEdges)-1)
		for i := 0; i < len(collisionVertexes)-1; i++ {
			middleVertex := Vx((collisionVertexes[i].X+collisionVertexes[i+1].X)/2, collisionVertexes[i].Y)

			for _, t := range mesh.finalMesh.Triangles {
				if t.HasVertex(middleVertex) {
					collisionTriangles[i] = CollisionTriangle{
						startVertex: collisionVertexes[i],
						endVertex:   collisionVertexes[i+1],
						startEdge:   collisionEdges[i],
						endEdge:     collisionEdges[i+1],
						triangle:    t,
					}
					break
				}
			}
		}

		curr := 0
		for x := dstBounds.Min.X; x <= dstBounds.Max.X; x++ {
			X := float64(x)

			// find collision triangle for the current point
			for i := curr; i < len(collisionTriangles); i++ {
				if collisionTriangles[i].startVertex.X <= X && X <= collisionTriangles[i].endVertex.X {
					curr = i
					break
				}
			}

			cTri := collisionTriangles[curr]
			beta := (X - cTri.startVertex.X) / (cTri.endVertex.X - cTri.startVertex.X)
			alphaStart := math.Abs(cTri.startEdge.Start.Y-Y) / math.Abs(cTri.startEdge.End.Y-cTri.startEdge.Start.Y)
			alphaEnd := math.Abs(cTri.endEdge.Start.Y-Y) / math.Abs(cTri.endEdge.End.Y-cTri.endEdge.Start.Y)

			origTri := mesh.primitiveMapping[cTri.triangle]
			lerpVertex := func(a, b Vertex, t float64) Vertex {
				origX := utils.LERP(a.X, b.X, t)
				origY := utils.LERP(a.Y, b.Y, t)

				return Vx(origX, origY)
			}

			origStartEdge := Edge{}
			origEndEdge := Edge{}
			for i, p := range origTri.points {
				if cTri.triangle.points[i].Equal(cTri.startEdge.Start) {
					origStartEdge.Start = p
				}

				if cTri.triangle.points[i].Equal(cTri.startEdge.End) {
					origStartEdge.End = p
				}

				if cTri.triangle.points[i].Equal(cTri.endEdge.Start) {
					origEndEdge.Start = p
				}

				if cTri.triangle.points[i].Equal(cTri.endEdge.End) {
					origEndEdge.End = p
				}
			}
			origStartVertex := lerpVertex(origStartEdge.Start, origStartEdge.End, alphaStart)
			origEndVertex := lerpVertex(origEndEdge.Start, origEndEdge.End, alphaEnd)
			origVertex := lerpVertex(origStartVertex, origEndVertex, beta)

			ret.(draw.Image).Set(x, y, mesh.initialMesh.Texture.At(int(math.Round(origVertex.X)), int(math.Round(origVertex.Y))))
		}
	}

	return ret
}
