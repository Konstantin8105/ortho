package ortho

import (
	"fmt"
	"io/ioutil"
	"os"
)

func view(
	points [][3]uint64,
	rectangles []Rectangle,
) {
	fmt.Fprintf(os.Stdout, "Points\n")
	fmt.Fprintf(os.Stdout, "%3s %5s %5s %5s\n",
		"ID", "X", "Y", "Z")
	for index, p := range points {
		fmt.Fprintf(os.Stdout, "%3d %5d %5d %5d\n",
			index, p[0], p[1], p[2])
	}

	fmt.Fprintf(os.Stdout, "Rectangles\n")
	fmt.Fprintf(os.Stdout, "%3s %3s %3s %3s %3s %12s\n",
		"ID", "P1", "P2", "P3", "P4", "Material")
	for index, r := range rectangles {
		fmt.Fprintf(os.Stdout, "%3d %3d %3d %3d %3d %12s\n",
			index,
			r.PointsId[0], r.PointsId[1], r.PointsId[2], r.PointsId[3],
			r.Material,
		)
	}
}

// debug create model in msh file format of gmsh
func debug(
	points [][3]uint64,
	rectangles []Rectangle,
) {

	var out string

	out += fmt.Sprintf(`$MeshFormat
2.2 0 8
$EndMeshFormat
$Nodes
`)
	out += fmt.Sprintf("%d\n", len(points))
	for index, p := range points {
		out += fmt.Sprintf("%3d %5d %5d %5d\n",
			index+1, p[0], p[1], p[2])
	}
	out += fmt.Sprintf(`$EndNodes
$Elements
`)

	out += fmt.Sprintf("%d\n", len(rectangles))
	for index, r := range rectangles {
		out += fmt.Sprintf("%3d 3 0 %3d %3d %3d %3d\n",
			index+1,
			r.PointsId[0]+1, r.PointsId[1]+1, r.PointsId[2]+1, r.PointsId[3]+1,
		)
	}
	out += fmt.Sprintf(`$EndElements`)

	err := ioutil.WriteFile("debug.msh", []byte(out), 0644)
	if err != nil {
		panic(err)
	}
}

func Example() {
	var m Model
	{
		fmt.Fprintf(os.Stdout, "Plate\n")
		m.Init(1800, 1200, "base")
		view(m.Generate(0))
		fmt.Fprintf(os.Stdout, "\n")
	}
	{
		fmt.Fprintf(os.Stdout, "Plate with split\n")
		m.Init(1800, 1200, "base")
		view(m.Generate(1000))
		fmt.Fprintf(os.Stdout, "\n")
	}
	{
		fmt.Fprintf(os.Stdout, "Horizontal\n")
		m.Init(1800, 1200, "base")
		m.Add(100, "stiff", 600, true)
		view(m.Generate(0))
		fmt.Fprintf(os.Stdout, "\n")
	}
	{
		fmt.Fprintf(os.Stdout, "Vertical\n")
		m.Init(1800, 1200, "base")
		m.Add(100, "stiff", 1000, false)
		view(m.Generate(0))
		fmt.Fprintf(os.Stdout, "\n")
	}
	{
		fmt.Fprintf(os.Stdout, "Horizontal and Vertical\n")
		m.Init(1800, 1200, "base")
		m.Add(100, "horizontal", 600, true)
		m.Add(100, "vertical", 1000, false)
		view(m.Generate(0))
		fmt.Fprintf(os.Stdout, "\n")
	}

	// Output:
	// Plate
	// Points
	//  ID     X     Y     Z
	//   0     0     0     0
	//   1     0  1200     0
	//   2  1800  1200     0
	//   3  1800     0     0
	// Rectangles
	//  ID  P1  P2  P3  P4     Material
	//   0   0   1   2   3         base
	//
	// Plate with split
	// Points
	//  ID     X     Y     Z
	//   0     0     0     0
	//   1     0   600     0
	//   2   900   600     0
	//   3   900     0     0
	//   4     0  1200     0
	//   5   900  1200     0
	//   6  1800   600     0
	//   7  1800     0     0
	//   8  1800  1200     0
	// Rectangles
	//  ID  P1  P2  P3  P4     Material
	//   0   0   1   2   3         base
	//   1   1   4   5   2         base
	//   2   3   2   6   7         base
	//   3   2   5   8   6         base
	//
	// Horizontal
	// Points
	//  ID     X     Y     Z
	//   0     0   600     0
	//   1     0   600   100
	//   2  1800   600   100
	//   3  1800   600     0
	//   4     0     0     0
	//   5  1800     0     0
	//   6     0  1200     0
	//   7  1800  1200     0
	// Rectangles
	//  ID  P1  P2  P3  P4     Material
	//   0   0   1   2   3        stiff
	//   1   4   0   3   5         base
	//   2   0   6   7   3         base
	//
	// Vertical
	// Points
	//  ID     X     Y     Z
	//   0  1000     0     0
	//   1  1000  1200     0
	//   2  1000  1200   100
	//   3  1000     0   100
	//   4     0     0     0
	//   5     0  1200     0
	//   6  1800  1200     0
	//   7  1800     0     0
	// Rectangles
	//  ID  P1  P2  P3  P4     Material

	//   0   0   1   2   3        stiff
	//   1   4   5   1   0         base
	//   2   0   1   6   7         base
	//
	// Horizontal and Vertical
	// Points
	//  ID     X     Y     Z
	//   0  1000     0     0
	//   1  1000   600     0
	//   2  1000   600   100
	//   3  1000     0   100
	//   4  1000  1200     0
	//   5  1000  1200   100
	//   6     0   600     0
	//   7     0   600   100
	//   8  1800   600   100
	//   9  1800   600     0
	//  10     0  1200     0
	//  11  1800  1200     0
	//  12     0     0     0
	//  13  1800     0     0
	// Rectangles
	//  ID  P1  P2  P3  P4     Material
	//   0   0   1   2   3     vertical
	//   1   1   4   5   2     vertical
	//   2   6   7   2   1   horizontal
	//   3   1   2   8   9   horizontal
	//   4   6  10   4   1         base
	//   5   1   4  11   9         base
	//   6  12   6   1   0         base
	//   7   0   1   9  13         base
}

func ExampleSelect() {
	var m Model
	m.Init(1800, 1200, "base")
	m.Add(100, "horizontal", 600, true)
	m.Add(100, "vertical", 1000, false)
	ps, _ := m.Generate(0)
	ts := Select(ps)
	fmt.Fprintf(os.Stdout, "%3s %5s %5s %5s\n",
		"ID", "X", "Y", "Z")
	for i := range ps {
		fmt.Fprintf(os.Stdout, "%3d %5d %5d %5d %s\n", i,
			ps[i][0], ps[i][1], ps[i][2],
			ts[i],
		)
	}
	// Output:
	// ID     X     Y     Z
	//   0  1000     0     0 Bottom
	//   1  1000   600     0 MainPlate
	//   2  1000   600   100 Other
	//   3  1000     0   100 Other
	//   4  1000  1200     0 Top
	//   5  1000  1200   100 Other
	//   6     0   600     0 Left
	//   7     0   600   100 Other
	//   8  1800   600   100 Other
	//   9  1800   600     0 Right
	//  10     0  1200     0 LeftTop
	//  11  1800  1200     0 RightTop
	//  12     0     0     0 LeftBottom
	//  13  1800     0     0 RightBottom
}

func ExampleAddPlateOnZ() {
	var m Model
	m.Init(1800, 1200, "base")
	m.Add(100, "horizontal", 600, true)
	m.Add(100, "vertical", 1000, false)
	m.AddPlateOnZ(150, 100, "horPlate", 600-75, true)
	m.AddPlateOnZ(50, 100, "vertPlate", 1000-25, false)
	ps, rs := m.Generate(0)
	debug(ps, rs)
	view(ps, rs)

	// Output:
	// Points
	//  ID     X     Y     Z
	//   0  1000     0     0
	//   1  1000   525     0
	//   2  1000   525   100
	//   3  1000     0   100
	//   4  1000   600     0
	//   5  1000   600   100
	//   6   975     0   100
	//   7   975   525   100
	//   8   975   600   100
	//   9  1025   525   100
	//  10  1025     0   100
	//  11  1025   600   100
	//  12  1000   675     0
	//  13  1000   675   100
	//  14  1000  1200     0
	//  15  1000  1200   100
	//  16   975   675   100
	//  17   975  1200   100
	//  18  1025   675   100
	//  19  1025  1200   100
	//  20     0   600     0
	//  21     0   600   100
	//  22   975   600     0
	//  23     0   525   100
	//  24     0   675   100
	//  25     0     0     0
	//  26     0   525     0
	//  27   975   525     0
	//  28   975     0     0
	//  29     0   675     0
	//  30   975   675     0
	//  31     0  1200     0
	//  32   975  1200     0
	//  33  1025   600     0
	//  34  1800   600   100
	//  35  1800   600     0
	//  36  1800   675   100
	//  37  1025   525     0
	//  38  1025     0     0
	//  39  1800   525     0
	//  40  1800     0     0
	//  41  1025   675     0
	//  42  1800   675     0
	//  43  1800   525   100
	//  44  1025  1200     0
	//  45  1800  1200     0
	// Rectangles
	//  ID  P1  P2  P3  P4     Material
	//   0   0   1   2   3     vertical
	//   1   1   4   5   2     vertical
	//   2   6   7   2   3    vertPlate
	//   3   7   8   5   2    vertPlate
	//   4   3   2   9  10    vertPlate
	//   5   2   5  11   9    vertPlate
	//   6   4  12  13   5     vertical
	//   7  12  14  15  13     vertical
	//   8   8  16  13   5    vertPlate
	//   9  16  17  15  13    vertPlate
	//  10   5  13  18  11    vertPlate
	//  11  13  15  19  18    vertPlate
	//  12  20  21   8  22   horizontal
	//  13  22   8   5   4   horizontal
	//  14  23  21   8   7     horPlate
	//  15  21  24  16   8     horPlate
	//  16  25  26  27  28         base
	//  17  28  27   1   0         base
	//  18  20  29  30  22         base
	//  19  22  30  12   4         base
	//  20  26  20  22  27         base
	//  21  27  22   4   1         base
	//  22  29  31  32  30         base
	//  23  30  32  14  12         base
	//  24   4   5  11  33   horizontal
	//  25  33  11  34  35   horizontal
	//  26  11  18  36  34     horPlate
	//  27   0   1  37  38         base
	//  28  38  37  39  40         base
	//  29   4  12  41  33         base
	//  30  33  41  42  35         base
	//  31   9  11  34  43     horPlate
	//  32   1   4  33  37         base
	//  33  37  33  35  39         base
	//  34  12  14  44  41         base
	//  35  41  44  45  42         base
}
