package avatar

import "fmt"

// const (
// 	cellSize = 16
// 	rows     = 16
// 	cols     = 16
// 	totalOps = 5
// )

const (
	cellSize = 2
	rows     = 128
	cols     = 128
)

type Color struct {
	R, G, B int
	Count   int
}

func pointSide(x, y, x1, y1, x2, y2 int) bool {
	return (x2-x1)*(y-y1)-(y2-y1)*(x-x1) < 0
}

func norm(c int) int {
	if c < 32 {
		c = 32
	} else if c > 127 {
		c = 127
	}
	return (c - 32) * 255 / 95
}

func GenerateAvatar(key string) string {
	// Parse ASCII values
	ascii := make([]int, 44)
	for i := 0; i < 44; i++ {
		ascii[i] = int(key[i])
	}

	// Grid holds optional colors with blend count
	grid := [rows][cols]*Color{}

	// Quadrants: top-left, top-right, bottom-left, bottom-right
	quads := [4][4]int{
		{0, cols / 2, 0, rows / 2},
		{cols / 2, cols, 0, rows / 2},
		{0, cols / 2, rows / 2, rows},
		{cols / 2, cols, rows / 2, rows},
	}

	// Start with the first 40 values
	baseData := make([]int, 40)
	copy(baseData, ascii[:40])

	// Loop through 4 quadrants
	for q := 0; q < 4; q++ {
		xMin, xMax, yMin, yMax := quads[q][0], quads[q][1], quads[q][2], quads[q][3]

		// 5 ops per quadrant
		for op := 0; op < 5; op++ {
			opData := make([]int, 8)
			for j := 0; j < 8; j++ {
				opData[j] = baseData[j]
			}

			x1 := xMin + (opData[0]*(xMax-xMin))/128
			y1 := yMin + (opData[1]*(yMax-yMin))/128
			x2 := xMin + (opData[2]*(xMax-xMin))/128
			y2 := yMin + (opData[3]*(yMax-yMin))/128
			r := norm(opData[4])
			g := norm(opData[5])
			b := norm(opData[6])
			sideSelector := opData[4] < 63

			// Paint cells on selected side
			for y := yMin; y < yMax; y++ {
				for x := xMin; x < xMax; x++ {
					if pointSide(x, y, x1, y1, x2, y2) == sideSelector {
						if grid[y][x] == nil {
							grid[y][x] = &Color{R: r, G: g, B: b, Count: 1}
						} else {
							c := grid[y][x]
							c.R = (c.R*c.Count + r) / (c.Count + 1)
							c.G = (c.G*c.Count + g) / (c.Count + 1)
							c.B = (c.B*c.Count + b) / (c.Count + 1)
							c.Count++
						}
					}
				}
			}

			// Rotate baseData left by 1 for next op
			baseData = append(baseData[1:], baseData[0])
		}
	}

	// Background color using normalized mapping
	bgR := norm(ascii[40])
	bgG := norm(ascii[41])
	bgB := norm(ascii[42])
	bgColor := fmt.Sprintf("rgb(%d,%d,%d)", bgR, bgG, bgB)

	// Start SVG
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256">` + "\n"
	svg += fmt.Sprintf(`<rect x="0" y="0" width="256" height="256" fill="%s"/>`+"\n", bgColor)
	svg += `<defs>
	    <filter id="blur" x="0" y="0" width="100%" height="100%">
	      <feGaussianBlur stdDeviation="15" />
	    </filter>
	  </defs>
	  <g filter="url(#blur)">` + "\n"

	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			c := grid[y][x]
			fill := bgColor
			if c != nil {
				fill = fmt.Sprintf("rgba(%d,%d,%d,0.5)", c.R, c.G, c.B)
			}
			svg += fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s" />`+"\n",
				x*cellSize, y*cellSize, cellSize, cellSize, fill)
		}
	}

	svg += `</g></svg>`
	return (svg)
	// 	// Parse ASCII values
	// 	ascii := make([]int, 44)
	// 	for i := 0; i < 44; i++ {
	// 		ascii[i] = int(key[i])
	// 	}

	// 	// Grid holds optional colors with blend count
	// 	grid := [rows][cols]*Color{}

	// 	// Choose quadrant based on last character
	// 	quadIndex := ascii[43] % 4
	// 	xMin, xMax, yMin, yMax := 0, cols/2, 0, rows/2
	// 	switch quadIndex {
	// 	case 1:
	// 		xMin, xMax = cols/2, cols
	// 	case 2:
	// 		yMin, yMax = rows/2, rows
	// 	case 3:
	// 		xMin, xMax = cols/2, cols
	// 		yMin, yMax = rows/2, rows
	// 	}

	// 	for i := 0; i < totalOps; i++ {
	// 		base := (i * 8) % 44
	// 		opData := make([]int, 8)
	// 		for j := 0; j < 8; j++ {
	// 			opData[j] = ascii[(base+j)%44]
	// 		}

	// 		x1 := xMin + (opData[0]*(xMax-xMin))/128
	// 		y1 := yMin + (opData[1]*(yMax-yMin))/128
	// 		x2 := xMin + (opData[2]*(xMax-xMin))/128
	// 		y2 := yMin + (opData[3]*(yMax-yMin))/128
	// 		r := norm(opData[4])
	// 		g := norm(opData[5])
	// 		b := norm(opData[6])

	// 		// Count cells on each side of the line
	// 		leftCount, rightCount := 0, 0
	// 		sideMap := [rows][cols]bool{}
	// 		for y := yMin; y < yMax; y++ {
	// 			for x := xMin; x < xMax; x++ {
	// 				right := pointSide(x, y, x1, y1, x2, y2)
	// 				sideMap[y][x] = right
	// 				if right {
	// 					rightCount++
	// 				} else {
	// 					leftCount++
	// 				}
	// 			}
	// 		}

	// 		smallerSide := false
	// 		if rightCount < leftCount {
	// 			smallerSide = true
	// 		}

	// 		for y := yMin; y < yMax; y++ {
	// 			for x := xMin; x < xMax; x++ {
	// 				if sideMap[y][x] == smallerSide {
	// 					coords := [][2]int{
	// 						{x, y}, {cols - 1 - x, y}, {x, rows - 1 - y}, {cols - 1 - x, rows - 1 - y},
	// 					}
	// 					for _, pos := range coords {
	// 						xp, yp := pos[0], pos[1]
	// 						if grid[yp][xp] == nil {
	// 							grid[yp][xp] = &Color{R: r, G: g, B: b, Count: 1}
	// 						} else {
	// 							c := grid[yp][xp]
	// 							c.R = (c.R*c.Count + r) / (c.Count + 1)
	// 							c.G = (c.G*c.Count + g) / (c.Count + 1)
	// 							c.B = (c.B*c.Count + b) / (c.Count + 1)
	// 							c.Count++
	// 						}
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}

	// 	// Background color using normalized mapping
	// 	bgR := norm(ascii[40])
	// 	bgG := norm(ascii[41])
	// 	bgB := norm(ascii[42])
	// 	bgColor := fmt.Sprintf("rgb(%d,%d,%d)", bgR, bgG, bgB)

	// 	// Start SVG
	// 	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256">` + "\n"
	// 	svg += fmt.Sprintf(`<rect x="0" y="0" width="256" height="256" fill="%s"/>`+"\n", bgColor)
	// 	svg += `<defs>
	//     <filter id="blur" x="0" y="0" width="100%" height="100%">
	//       <feGaussianBlur stdDeviation="0" />
	//     </filter>
	//   </defs>
	//   <g filter="url(#blur)">` + "\n"

	// 	for y := 0; y < rows; y++ {
	// 		for x := 0; x < cols; x++ {
	// 			c := grid[y][x]
	// 			fill := bgColor
	// 			if c != nil {
	// 				fill = fmt.Sprintf("rgba(%d,%d,%d,1)", c.R, c.G, c.B)
	// 			}
	// 			svg += fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s" />`+"\n",
	// 				x*cellSize, y*cellSize, cellSize, cellSize, fill)
	// 		}
	// 	}

	// 	svg += `</g></svg>`
	// 	return svg
}
