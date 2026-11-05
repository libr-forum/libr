package avatar

import "fmt"

const (
	cellSize = 32
	rows     = 8
	cols     = 8
	totalOps = 5
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
	ascii := make([]int, 44)
	for i := 0; i < 44; i++ {
		ascii[i] = int(key[i])
	}

	grid := [rows][cols]*Color{}

	quadIndex := ascii[43] % 4
	xMin, xMax, yMin, yMax := 0, cols/2, 0, rows/2
	switch quadIndex {
	case 1:
		xMin, xMax = cols/2, cols
	case 2:
		yMin, yMax = rows/2, rows
	case 3:
		xMin, xMax = cols/2, cols
		yMin, yMax = rows/2, rows
	}

	for i := 0; i < totalOps; i++ {
		base := (i * 8) % 44
		opData := make([]int, 8)
		for j := 0; j < 8; j++ {
			opData[j] = ascii[(base+j)%44]
		}

		x1 := xMin + (opData[0]*(xMax-xMin))/128
		y1 := yMin + (opData[1]*(yMax-yMin))/128
		x2 := xMin + (opData[2]*(xMax-xMin))/128
		y2 := yMin + (opData[3]*(yMax-yMin))/128
		r := norm(opData[4])
		g := norm(opData[5])
		b := norm(opData[6])

		leftCount, rightCount := 0, 0
		sideMap := [rows][cols]bool{}
		for y := yMin; y < yMax; y++ {
			for x := xMin; x < xMax; x++ {
				right := pointSide(x, y, x1, y1, x2, y2)
				sideMap[y][x] = right
				if right {
					rightCount++
				} else {
					leftCount++
				}
			}
		}

		smallerSide := false
		if rightCount < leftCount {
			smallerSide = true
		}

		for y := yMin; y < yMax; y++ {
			for x := xMin; x < xMax; x++ {
				if sideMap[y][x] == smallerSide {
					coords := [][2]int{
						{x, y}, {cols - 1 - x, y}, {x, rows - 1 - y}, {cols - 1 - x, rows - 1 - y},
					}
					for _, pos := range coords {
						xp, yp := pos[0], pos[1]
						if grid[yp][xp] == nil {
							grid[yp][xp] = &Color{R: r, G: g, B: b, Count: 1}
						} else {
							c := grid[yp][xp]
							c.R = (c.R*c.Count + r) / (c.Count + 1)
							c.G = (c.G*c.Count + g) / (c.Count + 1)
							c.B = (c.B*c.Count + b) / (c.Count + 1)
							c.Count++
						}
					}
				}
			}
		}
	}

	bgR := norm(ascii[40])
	bgG := norm(ascii[41])
	bgB := norm(ascii[42])
	bgColor := fmt.Sprintf("rgb(%d,%d,%d)", bgR, bgG, bgB)

	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256">` + "\n"
	svg += fmt.Sprintf(`<rect x="0" y="0" width="256" height="256" fill="%s"/>`+"\n", bgColor)
	svg += "<defs>\n    <filter id=\"blur\" x=\"0\" y=\"0\" width=\"100%\" height=\"100%\">\n      <feGaussianBlur stdDeviation=\"0\" />\n    </filter>\n  </defs>\n  <g filter=\"url(#blur)\">\n"

	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			c := grid[y][x]
			fill := bgColor
			if c != nil {
				fill = fmt.Sprintf("rgba(%d,%d,%d,1)", c.R, c.G, c.B)
			}
			svg += fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s" />`+"\n",
				x*cellSize, y*cellSize, cellSize, cellSize, fill)
		}
	}

	svg += `</g></svg>`
	return svg
}
