package main

// the best color ever: color.NRGBA{0x80, 0x80, 0xff, 0x80}

import(
	"log"
	"fmt"
	"math"
	"unsafe"
	"image/color"
	"github.com/hajimehoshi/ebiten"
    "github.com/hajimehoshi/ebiten/ebitenutil" // This is required to draw debug texts.
)

type object struct{
	shapeOffset vec3d
	rX, rY, rZ float64
	m, mesh matrix
}

// 2d vector
type vec2d struct {
	x, y float32
}

// 3d vector
type vec3d struct {
	x, y, z float32
}

// 2x2 matrix
type matrix2d struct {
	a [2]vec2d
}

// 3x3 matrix
type matrix3d struct {
	a [3]vec3d
}

// custom size matrix
type matrix struct {
	a [][]float32
}

// sinewave function for simplification of code(plus it converts the output from float64 to float32)
func sin(x float64) float32{
	return float32(math.Sin(x))
}

// cosinewave function for simplification of code(plus it converts the output from float64 to float32)
func cos(x float64) float32{
	return float32(math.Cos(x))
}

// multiply matrix2d by a number
func (A *matrix2d) mulNum(num float32) matrix2d{
	out := matrix2d{}
	for i, v := range A.a{
		out.a[i].x = v.x*num
		out.a[i].y = v.y*num
	}
	return out
}

// multiply matrix3d by a number
func (A *matrix3d) mulNum(num float32) matrix3d{
	out := matrix3d{}
	for i, v := range A.a{
		out.a[i].x = v.x*num
		out.a[i].y = v.y*num
		out.a[i].z = v.z*num
	}
	return out
}

// my matrix2d by matrix2d multiplication function(although mulMat is better)
func (A *matrix2d) mulMatrix2d(B matrix2d) matrix2d{
	out := matrix2d{}
	out.a[0].x = A.a[0].x*B.a[0].x+A.a[0].y*B.a[1].x
	out.a[0].y = A.a[0].x*B.a[0].y+A.a[0].y*B.a[1].y
	out.a[1].x = A.a[1].x*B.a[0].x+A.a[1].y*B.a[1].x
	out.a[1].y = A.a[1].x*B.a[0].y+A.a[1].y*B.a[1].y
	return out
}

func (A *matrix) mulNum(n1, n2 float32){
	for _, v := range A.a{
		v[0] *= n1
		v[1] *= n2
	}
}

func (A *matrix) mulAdd2(n1, n2 float32){
	for _, v := range A.a{
		v[0] += n1
		v[1] += n2
	}
}

func (A *matrix) mulAdd3(n1, n2, n3 float32){
	for _, v := range A.a{
		v[0] += n1
		v[1] += n2
		v[2] += n3
	}
}

// matrix multiplication function
func mulMat(x, y matrix) matrix {
	out := make([][]float32, len(x.a))
	for i := 0; i < len(x.a); i++ {
		out[i] = make([]float32, len(y.a[0]))
		for j := 0; j < len(y.a[0]); j++ {
			for k := 0; k < len(y.a); k++ {
				out[i][j] += x.a[i][k] * y.a[k][j]
			}
		}
	}
	return matrix{a:out}
}

// convert matrix2d to a matrix
func (A *matrix2d) matrix() matrix{
	out := make([][]float32, 2)
	for i := range out{
		out[i] = make([]float32, 2)
		out[i][0] = A.a[i].x
		out[i][1] = A.a[i].y
	}
	return matrix{a:out}
}

// make a new matrix with width w and height h
func newMatrix(w, h int) matrix{
	out := make([][]float32, w)
	for i := range out{
		out[i] = make([]float32, h)
	}
	return matrix{a: out}
}

func newTriangle(screen *ebiten.Image, x1, y1, x2, y2, x3, y3 float32, color color.NRGBA){
	drawLine(screen, x1, y1, x2, y2, color)
	drawLine(screen, x2, y2, x3, y3, color)
	drawLine(screen, x3, y3, x1, y1, color)
}

func drawLine(screen *ebiten.Image, x1, y1, x2, y2 float32, color color.NRGBA){
	ebitenutil.DrawLine(screen, float64(x1), float64(y1), float64(x2), float64(y2), color)
}

func fillTriangle(screen *ebiten.Image, x1, y1, x2, y2, x3, y3 float32, color color.NRGBA){
	// get length of all sides
	d1 := math.Sqrt(float64(((y2-y1)*(y2-y1))+((x2-x1)*(x2-x1))))
	d2 := math.Sqrt(float64(((y3-y2)*(y3-y2))+((x3-x2)*(x3-x2))))
	d3 := math.Sqrt(float64(((y1-y3)*(y1-y3))+((x1-x3)*(x1-x3))))
	counter := 0
	if d1<d2 || d1==d2 && d1<d2 || d1==d2{ // the first side is the shortest
		var tx float64 = float64(x1)
		var ty float64 = float64(y1)
		vx := float64(x2-x1)/d1
		vy := float64(y2-y1)/d1
		for float64(counter) < d1 {
			drawLine(screen, float32(x3), float32(y3), float32(tx), float32(ty), color)
			// drawing a line from point(x3,y3) to point(tx,ty).
			tx = float64(tx) + vx
			ty = float64(ty) + vy
			counter = counter + 1
		} 
	}else if d2<d3 || d2==d3{ // the second side is the shortest
		var tx float64 = float64(x2)
		var ty float64 = float64(y2)
		vx := float64(x3-x2)/d2
		vy := float64(y3-y2)/d2
		for float64(counter)<d2 {
			drawLine(screen, float32(x1), float32(y1), float32(tx), float32(ty), color)
			tx = tx + vx
			ty = ty + vy
			counter = counter + 1
		}
	}else{ // the third side is shortest
		var tx float64 = float64(x3)
		var ty float64 = float64(y3)
		vx := float64(x1-x3)/d3
		vy := float64(y1-y3)/d3
		for float64(counter)<d3 {
			drawLine(screen, float32(x2), float32(y2), float32(tx), float32(ty), color)
			tx = tx + vx
			ty = ty + vy
			counter = counter + 1
		}
	}

}

// the quick reverse square root of x (1/sqrt(x)) from quake 2 implemetation in golang, made by me
func Q_rsqrt(x float32) float32{
	var i int32
	var x2, y float32
	const threehalfs = 1.5
	
	x2 = x*0.5
	y = x
	i = *(*int32)(unsafe.Pointer(&y)) // evil floating point bit hack
	i = 0x5f3759df - (i>>1) // what the fuck?
	y = *(*float32)(unsafe.Pointer(&i))
	y = y * (threehalfs - (x2*y*y)) // 1st iteration
//	y = y * (threehalfs - (x2*y*y)) // 2nd iteration, this can be removed
	
	return y
}

// some variables necessary for the rendering
var shapeOffset, light_direction vec3d
var w, h int
var rX, rY, rZ float64
var m, mesh, matproj matrix
var vCamera vec3d
var cR, cG, cB float32 
var cA uint8 

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func update(screen *ebiten.Image) error {
	if ebiten.IsDrawingSkipped() {
		return nil
	}
	screen.Clear()
	
	// setup and calculate the rotation matrices
	rotateX := newMatrix(3, 3)
	rotateX.a[0] = []float32{1, 0, 0}
	rotateX.a[1] = []float32{0, cos(rX), -sin(rX)}
	rotateX.a[2] = []float32{0, sin(rX), cos(rX)}
	rotateY := newMatrix(3, 3)
	rotateY.a[0] = []float32{cos(rY), 0, -sin(rY)}
	rotateY.a[1] = []float32{0, 1, 0}
	rotateY.a[2] = []float32{sin(rY), 0, cos(rY)}
	rotateZ := newMatrix(3, 3)
	rotateZ.a[0] = []float32{cos(rZ), -sin(rZ), 0}
	rotateZ.a[1] = []float32{sin(rZ), cos(rZ), 0}
	rotateZ.a[2] = []float32{0, 0, 1}
	
	// update the rotation angles
	rX += 0.01
	rY += 0.0
	rZ += 0.01
	
	// multiply the shape by the rotation matrices
	x := mulMat(m, rotateX)
	x = mulMat(x, rotateY)
	x = mulMat(x, rotateZ)
	x.mulAdd3(shapeOffset.x, shapeOffset.y, shapeOffset.z)
	
	// do some division by z to get perspective
	for i, v := range x.a{
		x.a[i][0] = v[0]/v[2]
		x.a[i][1] = v[1]/v[2]
	}
	
	// setup our crude render matrix
	n := newMatrix(3, 2)
	n.a[0] = []float32{1, 0}
	n.a[1] = []float32{0, 1}
	
	// render it out as a complete shape rather than just plain points
	fmt.Println("NORMALS: ")
	for _, v := range mesh.a{
		var normal, line1, line2 vec3d
		line1.x = x.a[int(v[1]-1)][0] - x.a[int(v[0]-1)][0]
		line1.y = x.a[int(v[1]-1)][1] - x.a[int(v[0]-1)][1]
		line1.z = x.a[int(v[1]-1)][2] - x.a[int(v[0]-1)][2]
		
		line2.x = x.a[int(v[2]-1)][0] - x.a[int(v[0]-1)][0]
		line2.y = x.a[int(v[2]-1)][1] - x.a[int(v[0]-1)][1]
		line2.z = x.a[int(v[2]-1)][2] - x.a[int(v[0]-1)][2]
		
		normal.x = line1.y * line2.z - line1.z * line2.y
		normal.y = line1.z * line2.x - line1.x * line2.z
		normal.z = line1.x * line2.y - line1.y * line2.x
		
		var l float32 = Q_rsqrt(normal.x*normal.x + normal.y*normal.y + normal.z*normal.z)
		normal.x *= l
		normal.y *= l
		normal.z *= l
		var x1, x2, x3, y1, y2, y3 float32
		x1 = x.a[int(v[0]-1)][0]*(float32(w)/8)+(float32(w)/2)
		y1 = x.a[int(v[0]-1)][1]*(float32(h)/8)+(float32(h)/2)
		x2 = x.a[int(v[1]-1)][0]*(float32(w)/8)+(float32(w)/2)
		y2 = x.a[int(v[1]-1)][1]*(float32(h)/8)+(float32(h)/2)
		x3 = x.a[int(v[2]-1)][0]*(float32(w)/8)+(float32(w)/2)
		y3 = x.a[int(v[2]-1)][1]*(float32(h)/8)+(float32(h)/2)
		//var dP float32 = normal.x*(x.a[int(v[0]-1)][0] - vCamera.x) + normal.y*(x.a[int(v[0]-1)][1] - vCamera.y) + normal.z*(x.a[int(v[0]-1)][2] - vCamera.z)
		if(normal.z<0){
			
			// Illumination
			l = Q_rsqrt(light_direction.x*light_direction.x + light_direction.y*light_direction.y + light_direction.z*light_direction.z)
			light_direction.x *= l
			light_direction.y *= l
			light_direction.z *= l
			
			// How similar is normal to light direction
			dp := normal.x * light_direction.x + normal.y * light_direction.y + normal.z * light_direction.z
			fmt.Println(dp)
			c := color.NRGBA{uint8(cR*0.1), uint8(cG*0.1), uint8(cB*0.1), cA}
			if dp>0.1{
				c = color.NRGBA{uint8(cR*dp), uint8(cG*dp), uint8(cB*dp), cA}
			}
			if 1==0{fmt.Println(c)}
			fillTriangle(screen, x1, y1, x2, y2, x3, y3, c)
		}
	}
    return nil
}

func main(){
	w, h = 800, 600
	n := float64(1)
	shapeOffset = vec3d{0, 0, 1}
	
	// specify the light direction
	light_direction = vec3d{ -1, 0, -.5 }
	
	// setup rotation angles for the shape
	rX = 0
	rY = 0
	rZ = 0
	
	// specify the color of our cube
	cR = 128
	cG = 128
	cB = 255
	cA = 0xff
	
	// setup the shape vertices themselves(cube shape in this case, center in the middle)
	m = newMatrix(9, 3)
	m.a[0] = []float32{-0.5, -0.5, -0.5}
	m.a[1] = []float32{0.5, -0.5, -0.5}
	m.a[2] = []float32{-0.5, 0.5, -0.5}
	m.a[3] = []float32{0.5, 0.5, -0.5}
	m.a[4] = []float32{-0.5, -0.5, 0.5}
	m.a[5] = []float32{0.5, -0.5, 0.5}
	m.a[6] = []float32{-0.5, 0.5, 0.5}
	m.a[7] = []float32{0.5, 0.5, 0.5}
	//m.a[8] = []float32{0, 1.5, 0}
	
	// setup the mesh of the shape(used for rendering it out as a whole shape, not only the points)
	mesh = newMatrix(12, 3)
	mesh.a[0] = []float32{2, 1, 3}
	mesh.a[1] = []float32{3, 4, 2}
	mesh.a[2] = []float32{4, 3, 7}
	mesh.a[3] = []float32{7, 8, 4}
	mesh.a[4] = []float32{2, 4, 8}
	mesh.a[5] = []float32{8, 6, 2}
	mesh.a[6] = []float32{1, 6, 5}
	mesh.a[7] = []float32{6, 1, 2}
	mesh.a[8] = []float32{8, 7, 5}
	mesh.a[9] = []float32{5, 6, 8}
	mesh.a[10] = []float32{5, 7, 3}
	mesh.a[11] = []float32{3, 1, 5}
	
	if err := ebiten.Run(update, w, h, n, "ebiten3D"); err != nil {
		log.Fatal(err)
	}
}
