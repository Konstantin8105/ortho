package ortho

import (
	"sort"
)

type Model struct {
	width, heigth uint64 // mm
	plates        []Plate
	cuts          []cut
}

func (m Model) copy() (c Model) {
	c.width, c.heigth = m.width, m.heigth
	c.plates = make([]Plate, len(m.plates))
	copy(c.plates, m.plates)
	c.cuts = make([]cut, len(m.cuts))
	copy(c.cuts, m.cuts)
	return
}

func (m *Model) Init(W, H uint64, material string) {
	m.width = W
	m.heigth = H
	m.cuts = nil
	m.plates = nil
	m.plates = append(m.plates, Plate{
		Width:    W,
		Heigth:   H,
		Coord:    [2]uint64{0, 0},
		Plane:    XOY,
		Offset:   0,
		Material: material,
	})
}

type cut struct {
	offset uint64
	plane  Planes
}

func (m *Model) Add(stiffH uint64, material string, offset uint64, parallelX bool) {
	if parallelX {
		m.plates = append(m.plates, Plate{
			Width:    m.width,
			Heigth:   stiffH,
			Coord:    [2]uint64{0, 0},
			Plane:    ZOX,
			Offset:   offset,
			Material: material,
		})
		m.cuts = append(m.cuts, cut{plane: ZOX, offset: offset})
		if m.heigth < offset {
			panic("not valid offset")
		}
	} else {
		m.plates = append(m.plates, Plate{
			Width:    m.heigth,
			Heigth:   stiffH,
			Coord:    [2]uint64{0, 0},
			Plane:    YOZ,
			Offset:   offset,
			Material: material,
		})
		m.cuts = append(m.cuts, cut{plane: YOZ, offset: offset})
		if m.width < offset {
			panic("not valid offset")
		}
	}
	m.cuts = append(m.cuts, cut{plane: XOY, offset: 0})
	m.cuts = append(m.cuts, cut{plane: XOY, offset: stiffH})
}

// ParalleX = false:
//	Y
//	|    --- +-----------------------+
//	|     |  |                       |
//	|     W  |                       |
//	|     |  |                       |
//	|    --- +-----------------------+
//	|     |
//	|     |
//	|     Offset
//	|     |
//	|     |
//	*--------------------------------> X
//
//
// ParalleX = true:
//	               +-------+
//	               |       |
//	Y              |       |
//	|              |       |
//	|              |       |
//	               |       |
//	*              +-------+  --> X
//	|<-- Offset -->|<- W ->|
//
func (m *Model) AddPlateOnZ(W, Z uint64, material string, offset uint64, parallelX bool) {
	if parallelX {
		if m.width < offset+W {
			panic("not valid offset")
		}
		m.plates = append(m.plates, Plate{
			Width:    m.width,
			Heigth:   W,
			Coord:    [2]uint64{0, offset},
			Plane:    XOY,
			Offset:   Z,
			Material: material,
		})
		m.cuts = append(m.cuts, cut{plane: ZOX, offset: offset})
		m.cuts = append(m.cuts, cut{plane: ZOX, offset: offset + W})
	} else {
		if m.width < offset+W {
			panic("not valid offset")
		}
		m.plates = append(m.plates, Plate{
			Width:    W,
			Heigth:   m.heigth,
			Coord:    [2]uint64{offset, 0},
			Plane:    XOY,
			Offset:   Z,
			Material: material,
		})
		m.cuts = append(m.cuts, cut{plane: YOZ, offset: offset})
		m.cuts = append(m.cuts, cut{plane: YOZ, offset: offset + W})
	}
	m.cuts = append(m.cuts, cut{plane: XOY, offset: Z})
}

func (m Model) Generate(maxDistance uint64) (
	points [][3]uint64,
	rectangles []Rectangle,
) {
	// edit only copy of model
	m = m.copy()

	// cut all plates
	for i := range m.cuts {
		plane := m.cuts[i].plane
		offset := m.cuts[i].offset
		var app []Plate
	again:
		var cuted bool
		for i := 0; i < len(m.plates); i++ {
			ps, cuts := m.plates[i].cut(offset, plane)
			if !cuts {
				continue
			}
			app = append(app, ps[0], ps[1])
			m.plates = append(m.plates[:i], m.plates[i+1:]...)
			cuted = true
		}
		if cuted {
			goto again
		}
		m.plates = append(m.plates, app...)
	}

	var add func(v1, v2 uint64, plane Planes)

	// function cuts by maxDistance
	add = func(v1, v2 uint64, plane Planes) {
		if v2 < v1 {
			panic("not valid values")
		}
		if maxDistance < v2-v1 {
			mid := v1 + (v2-v1)/2
			m.cuts = append(m.cuts, cut{offset: mid, plane: plane})
			add(v1, mid, plane)
			add(mid, v2, plane)
		}
	}

	// split into small pieces in according to maxDistance
	if 0 < maxDistance {
		for _, p := range m.plates {
			switch p.Plane {
			case XOY:
				add(p.Coord[0], p.Coord[0]+p.Width, YOZ)
				add(p.Coord[1], p.Coord[1]+p.Heigth, ZOX)
			case YOZ:
				add(p.Coord[0], p.Coord[0]+p.Heigth, XOY)
			case ZOX:
				add(p.Coord[0], p.Coord[0]+p.Heigth, XOY)
			}
		}
		return m.Generate(0)
	}

	// prepare result data
	for _, p := range m.plates {
		// convert 2d points of plate into 3D implementation
		//	1-----2
		//	|     |
		//	0-----3
		var p3d [4][3]uint64
		switch p.Plane {
		case XOY:
			p3d = [4][3]uint64{
				{p.Coord[0], p.Coord[1], p.Offset},
				{p.Coord[0], p.Coord[1] + p.Heigth, p.Offset},
				{p.Coord[0] + p.Width, p.Coord[1] + p.Heigth, p.Offset},
				{p.Coord[0] + p.Width, p.Coord[1], p.Offset},
			}
		case ZOX:
			p3d = [4][3]uint64{
				{p.Coord[0], p.Offset, p.Coord[1]},
				{p.Coord[0], p.Offset, p.Coord[1] + p.Heigth},
				{p.Coord[0] + p.Width, p.Offset, p.Coord[1] + p.Heigth},
				{p.Coord[0] + p.Width, p.Offset, p.Coord[1]},
			}
		case YOZ:
			p3d = [4][3]uint64{
				{p.Offset, p.Coord[0], p.Coord[1]},
				{p.Offset, p.Coord[0] + p.Width, p.Coord[1]},
				{p.Offset, p.Coord[0] + p.Width, p.Coord[1] + p.Heigth},
				{p.Offset, p.Coord[0], p.Coord[1] + p.Heigth},
			}
		}

		var r Rectangle
		for i := range p3d {
			found := -1
			for j := range points {
				if points[j][0] == p3d[i][0] &&
					points[j][1] == p3d[i][1] &&
					points[j][2] == p3d[i][2] {
					found = j
					break
				}
			}
			if 0 <= found { // found point in points slice
				r.PointsId[i] = found
				continue
			}
			// new point
			points = append(points, p3d[i])
			r.PointsId[i] = len(points) - 1
		}

		r.Material = p.Material
		rectangles = append(rectangles, r)
	}

	{
		var xoyListOnZ []int
		for i := range rectangles {
			if points[rectangles[i].PointsId[0]][2] == 0 &&
				points[rectangles[i].PointsId[0]][2] ==
					points[rectangles[i].PointsId[2]][2] {
				continue
			}
			xoyListOnZ = append(xoyListOnZ, i)
		}
		removeList := []int{}
		for i := range xoyListOnZ {
			for j := range xoyListOnZ {
				if i <= j {
					continue
				}
				if rectangles[xoyListOnZ[i]].PointsId[0] ==
					rectangles[xoyListOnZ[j]].PointsId[0] {
					removeList = append(removeList, xoyListOnZ[i])
				}
			}
		}
		sort.Ints(removeList)
		for i := range removeList {
			ind := removeList[len(removeList)-i-1]
			rectangles = append(rectangles[:ind], rectangles[ind+1:]...)
		}
	}

	return
}

type Rectangle struct {
	PointsId [4]int // indexes of points
	Material string
}

type Plate struct {
	Width, Heigth uint64    // mm
	Coord         [2]uint64 // mm

	Plane  Planes
	Offset uint64

	Material string
}

func (p Plate) cut(d uint64, plane Planes) (out [2]Plate, cuts bool) {
	if p.Plane == plane {
		return
	}

	var vertical bool

	switch p.Plane {
	case XOY:
		switch plane {
		// no need : case XOY:
		case YOZ:
			vertical = true
		case ZOX:
		default:
			panic(plane)
		}
	case YOZ:
		switch plane {
		case XOY:
			// no need : case YOZ:
		case ZOX:
			vertical = true
		default:
			panic(plane)
		}
	case ZOX:
		switch plane {
		case XOY:
		case YOZ:
			vertical = true
		// no need : case ZOX:
		default:
			panic(plane)
		}
	default:
		panic(p.Plane)
	}

	if vertical {
		if p.Coord[0] < d && d < p.Coord[0]+p.Width {
			//	1---|--2
			//	|   |  |
			//	0---|--3
			cuts = true
			out[0].Coord = p.Coord
			out[0].Width = d - p.Coord[0]
			out[0].Heigth = p.Heigth

			out[1].Coord = [2]uint64{d, p.Coord[1]}
			out[1].Width = p.Coord[0] + p.Width - d
			out[1].Heigth = p.Heigth
		}
	} else {
		if p.Coord[1] < d && d < p.Coord[1]+p.Heigth {
			//	1------2
			//	|      |
			//	--------
			//	|      |
			//	0------3
			cuts = true
			out[0].Coord = p.Coord
			out[0].Width = p.Width
			out[0].Heigth = d - p.Coord[1]

			out[1].Coord = [2]uint64{p.Coord[0], d}
			out[1].Width = p.Width
			out[1].Heigth = p.Coord[1] + p.Heigth - d
		}
	}

	for i := range out {
		out[i].Plane = p.Plane
		out[i].Offset = p.Offset
		out[i].Material = p.Material
	}

	return
}

type Planes uint8

const (
	XOY Planes = iota + 1 // 1
	YOZ                   // 2
	ZOX                   // 3
)

func (p Planes) String() string {
	switch p {
	case XOY:
		return "XOY"
	case YOZ:
		return "YOZ"
	case ZOX:
		return "ZOX"
	}
	panic("undefine plane")
}

type PointType uint8

const (
	Other PointType = iota
	MainPlate
	Left
	Right
	Top
	Bottom
	LeftTop
	LeftBottom
	RightTop
	RightBottom
)

func (t PointType) String() string {
	switch t {
	case Other:
		return "Other"
	case MainPlate:
		return "MainPlate"
	case Left:
		return "Left"
	case Right:
		return "Right"
	case Top:
		return "Top"
	case Bottom:
		return "Bottom"
	case LeftTop:
		return "LeftTop"
	case LeftBottom:
		return "LeftBottom"
	case RightTop:
		return "RightTop"
	case RightBottom:
		return "RightBottom"
	}
	panic("undefined")
}

// return point indexes
func Select(points [][3]uint64) (types []PointType) {
	// dimensions
	var l, r, t, b uint64
	for i := range points {
		x, y := points[i][0], points[i][1]
		if x < l {
			l = x
		}
		if r < x {
			r = x
		}
		if t < y {
			t = y
		}
		if y < b {
			b = y
		}
	}

	// classification
	types = make([]PointType, len(points))
	for i := range points {
		if points[i][2] != 0 {
			continue
		}
		types[i] = MainPlate

		x, y := points[i][0], points[i][1]

		if x == l {
			types[i] = Left
		}
		if x == r {
			types[i] = Right
		}
		if y == t {
			types[i] = Top
		}
		if y == b {
			types[i] = Bottom
		}

		if x == l && y == b {
			types[i] = LeftBottom
		}
		if x == r && y == t {
			types[i] = RightTop
		}
		if x == l && y == t {
			types[i] = LeftTop
		}
		if x == r && y == b {
			types[i] = RightBottom
		}
	}
	return
}
