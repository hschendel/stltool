package main

import (
	"flag"
	"fmt"
	"os"
	
	"github.com/hschendel/stl"
)

const version = "1.0.1"

// Command line parameters
type Params struct {
	Command      string
	InputFile    string
	OutputFile   string
	OutputAscii  bool
	OutputBinary bool
	ShowMatrix   bool
	Fixnorm      bool
	Angle        float64
	PosX         float64
	PosY         float64
	PosZ         float64
	DirX         float64
	DirY         float64
	DirZ         float64
	Factor       float64
	Matrix       *stl.Mat4
}

// Context for main()
type Run struct {
	Params        Params
	Solid         *stl.Solid
	DoReadInput   bool
	DoWriteOutput bool
	Command       func(run *Run)
	ExitCode      int
}

func (run *Run) ParseParams() {
	p := &(run.Params)
	flag.StringVar(&p.InputFile, "i", "", "Input File, default is stdin")
	flag.StringVar(&p.OutputFile, "o", "", "Output File, default is stdout")
	flag.BoolVar(&p.OutputAscii, "ascii", false, "Write output file in human-readable STL ASCII format, not recommended")
	flag.BoolVar(&p.OutputBinary, "binary", false, "Write output file in STL binary format")
	flag.BoolVar(&p.ShowMatrix, "showmat", false, "Print 4x4 transformation matrix to stderr in format [[m11 m12 m13 m14] ... [m41 m42 m43 m44]]")
	flag.BoolVar(&p.Fixnorm, "fixnorm", false, "Recalculate normals, makes only sense together with copy")
	flag.Float64Var(&p.Angle, "a", 90, "Rotation angle in degree")
	flag.Float64Var(&p.PosX, "x", 0, "x for point on rotation axis for rotate, translation on x-axis for translate, x-size in fitbox")
	flag.Float64Var(&p.PosY, "y", 0, "y for point on rotation axis for rotate, translation on y-axis for translate, y-size in fitbox")
	flag.Float64Var(&p.PosZ, "z", 0, "z for point on rotation axis for rotate, translation on z-axis for translate, z-size in fitbox")
	flag.Float64Var(&p.DirX, "dx", 0, "x for direction of rotation axis")
	flag.Float64Var(&p.DirY, "dy", 0, "y for direction of rotation axis")
	flag.Float64Var(&p.DirZ, "dz", 0, "z for direction of rotation axis")
	flag.Float64Var(&p.Factor, "f", 1, "scale factor")
	var matrixStr string
	flag.StringVar(&matrixStr, "m", "[[1 0 0 0] [0 1 0 0] [0 0 1 0] [0 0 0 1]]", "Transformation matrix for transform, same format as printed using -showmat.  Default is the identity matrix.")
	flag.Parse()
	p.Matrix = run.parseMat4("-m", matrixStr)

	if flag.NArg() == 1 || flag.NArg() == 2 {
		p.Command = flag.Arg(0)
		if flag.NArg() == 2 {
			if p.InputFile != "" || p.OutputFile != "" {
				fmt.Fprintln(os.Stderr, "Error: providing a single filename after the command forbids the use of -i or -o.")
				p.Command = "help"
				run.ExitCode = 15
			} else {
				p.InputFile = flag.Arg(1)
				p.OutputFile = flag.Arg(1)
			}
		}
	} else if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Error: no command")
		p.Command = "help"
		run.ExitCode = 10
	} else if flag.NArg() > 2 {
		fmt.Fprintln(os.Stderr, "Error: Too many arguments")
		p.Command = "help"
		run.ExitCode = 13
	}
	if p.OutputAscii && p.OutputBinary {
		fmt.Fprintln(os.Stderr, "Error: -ascii and -binary are exclusive")
		p.Command = "help"
		run.ExitCode = 12
	}
}

func (run *Run) parseMat4(param, s string) *stl.Mat4 {
	if s == "" {
		return nil
	}
	var r stl.Mat4
	n, err := fmt.Sscanf(s, "[[%f %f %f %f] [%f %f %f %f] [%f %f %f %f] [%f %f %f %f]]",
		&r[0][0], &r[0][1], &r[0][2], &r[0][3],
		&r[1][0], &r[1][1], &r[1][2], &r[1][3],
		&r[2][0], &r[2][1], &r[2][2], &r[2][3],
		&r[3][0], &r[3][1], &r[3][2], &r[3][3],
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Cannot read %s - %s\n", param, err.Error())
		return nil
	} else if n == 0 {
		return nil
	} else {
		return &r
	}
}

func (run *Run) ReadInput() bool {
	var err error
	if run.Params.InputFile == "" {
		run.Solid, err = stl.ReadAll(os.Stdin)
	} else {
		run.Solid, err = stl.ReadFile(run.Params.InputFile)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		run.ExitCode = 40
		return false
	} else {
		return true
	}
}

func (run *Run) WriteOutput() bool {
	if run.Params.OutputAscii {
		run.Solid.IsAscii = true
	} else if run.Params.OutputBinary {
		run.Solid.IsAscii = false
	}

	var err error
	if run.Params.OutputFile == "" {
		err = run.Solid.WriteAll(os.Stdout)
	} else {
		err = run.Solid.WriteFile(run.Params.OutputFile)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		run.ExitCode = 50
		return false
	} else {
		return true
	}
}

func (run *Run) Help() {
	fmt.Fprintln(os.Stderr, "stltool v"+version+`
stltool can be used to measure, validate, and transform .stl files used
for 3D printing.

Usage:

  stltool [Options] <command> [filename]
  
  The command always has to be provided after the options.

List of commands:

  copy        Simply copy from input to output, can be used to convert from ASCII
              to binary or vice versa with -binary or -ascii
  fitbox      Linear scale down into box described by -x, -y, and -z
  help        This help
  license     Print license information
  measure     Print size measured on x, y, and z axes in format [x y z]
  rotate      Rotate around axis described by point -x, -y, -z and direction vector
              -dx, -dy, -dz for -a degree
  scale       Scale by -f
  transform   Custom transformation with matrix -m
  translate   Translate (i.e. move) by -x, -y, -z
  validate    Print violations of the STL format, such as holes in the model

List of options:
`)
	flag.PrintDefaults()
	run.DoWriteOutput = false
}

func (run *Run) Measure() {
	m := run.Solid.Measure()
	out := os.Stdout
	fmt.Fprintf(out, "X-Axis length %v, from %v to %v\n", m.Len[0], m.Min[0], m.Max[0])
	fmt.Fprintf(out, "Y-Axis length %v, from %v to %v\n", m.Len[1], m.Min[1], m.Max[1])
	fmt.Fprintf(out, "Z-Axis length %v, from %v to %v\n", m.Len[2], m.Min[2], m.Max[2])
}

func (run *Run) Scale() {
	run.Solid.Scale(run.Params.Factor)
	run.DoWriteOutput = true
	if run.Params.ShowMatrix {
		m := stl.Mat4{
			stl.Vec4{run.Params.Factor, 0, 0, 0},
			stl.Vec4{0, run.Params.Factor, 0, 0},
			stl.Vec4{0, 0, run.Params.Factor, 0},
			stl.Vec4{0, 0, 0, 1},
		}
		fmt.Fprintf(os.Stderr, "%v\n", m)
	}
}

func (run *Run) Translate() {
	run.Solid.Translate(stl.Vec3{float32(run.Params.PosX), float32(run.Params.PosY), float32(run.Params.PosZ)})
	run.DoWriteOutput = true
	if run.Params.ShowMatrix {
		m := stl.Mat4{
			stl.Vec4{1, 0, 0, run.Params.PosX},
			stl.Vec4{0, 1, 0, run.Params.PosY},
			stl.Vec4{0, 0, 1, run.Params.PosZ},
			stl.Vec4{0, 0, 0, 1},
		}
		fmt.Fprintf(os.Stderr, "%v\n", m)
	}
}

func (run *Run) Rotate() {
	if !(run.Params.DirX > 0 || run.Params.DirY > 0 || run.Params.DirZ > 0) {
		os.Stderr.WriteString("The rotation axis must have a direction vector set by -dx, -dy, -dz\n")
		run.ExitCode = 20
		return
	}
	angleRad := run.Params.Angle / 360 * stl.TwoPi
	run.Solid.Rotate(
		stl.Vec3{float32(run.Params.PosX), float32(run.Params.PosY), float32(run.Params.PosZ)},
		stl.Vec3{float32(run.Params.DirX), float32(run.Params.DirY), float32(run.Params.DirZ)},
		angleRad,
	)
	run.DoWriteOutput = true
	if run.Params.ShowMatrix {
		var m stl.Mat4
		stl.RotationMatrix(
			stl.Vec3{float32(run.Params.PosX), float32(run.Params.PosY), float32(run.Params.PosZ)},
			stl.Vec3{float32(run.Params.DirX), float32(run.Params.DirY), float32(run.Params.DirZ)},
			angleRad,
			&m)
		fmt.Fprintf(os.Stderr, "%v\n", m)
	}
}

func (run *Run) Transform() {
	if run.Params.Matrix == nil {
		fmt.Fprintln(os.Stderr, "transform not possible without correct -m")
		run.ExitCode = 22
		return
	}
	run.Solid.Transform(run.Params.Matrix)
	run.DoWriteOutput = true
	if run.Params.ShowMatrix {
		fmt.Fprintf(os.Stderr, "%v\n", *run.Params.Matrix)
	}
}

func (run *Run) Copy() {
	run.DoWriteOutput = true
}

func (run *Run) FitBox() {
	if !(run.Params.PosX > 0 && run.Params.PosY > 0 && run.Params.PosZ > 0) {
		fmt.Fprintln(os.Stderr, "The size box has to be defined with -x, -y, -z each being > 0")
		run.ExitCode = 21
		return
	}
	run.Solid.ScaleLinearDowntoSizeBox(stl.Vec3{float32(run.Params.PosX), float32(run.Params.PosY), float32(run.Params.PosZ)})
	run.DoWriteOutput = true
}

func (run *Run) Validate() {
	out := os.Stdout
	if !run.Solid.IsInPositive() {
		fmt.Fprintln(out, "Some vertices have negative coordinates.")
	}
	triangleErrorsMap := run.Solid.Validate()
	for triangleIndex := range run.Solid.Triangles { // to be ordered
		triangleErrors, hasErrors := triangleErrorsMap[triangleIndex]
		if !hasErrors {
			continue
		}
		bytePos := ""
		if !run.Solid.IsAscii {
			bytePos = fmt.Sprintf(" at byte #%d", 84+50*triangleIndex)
		}
		fmt.Fprintf(out, "In triangle #%d%s (Vertices %v %v %v):\n",
			triangleIndex, bytePos,
			run.Solid.Triangles[triangleIndex].Vertices[0],
			run.Solid.Triangles[triangleIndex].Vertices[1],
			run.Solid.Triangles[triangleIndex].Vertices[2])
		if triangleErrors.HasEqualVertices {
			fmt.Fprintln(out, "  - There are identical vertices, meaning this is no triangle.")
		}
		if triangleErrors.NormalDoesNotMatch {
			fmt.Fprintf(out, "  - The normal vector %v cannot be calculated from the vertices.\n")
		}
		for e := 0; e < 3; e++ {
			edgeError := triangleErrors.EdgeErrors[e]
			if edgeError != nil {
				fmt.Fprintf(out, "  - In edge #%d from %v to %v:\n", e,
					run.Solid.Triangles[triangleIndex].Vertices[e],
					run.Solid.Triangles[triangleIndex].Vertices[(e+1)%3])
				if edgeError.IsUsedInOtherTriangles() {
					if len(edgeError.SameEdgeTriangles) == 1 {
						fmt.Fprintf(out, "    - Edge is also used in triangle #%d.\n", edgeError.SameEdgeTriangles[0])
					} else {
						fmt.Fprintf(out, "    - Edge is also used in triangles with indices %v.\n", edgeError.SameEdgeTriangles)
					}
				}
				if edgeError.HasMultipleCounterEdges() {
					fmt.Fprintf(out, "    - The edge has more than one matching opposing edge in triangles with indices %v.\n", edgeError.CounterEdgeTriangles)
				} else if edgeError.HasNoCounterEdge() {
					fmt.Fprintf(out, "    - There is no triangle with opposing edge from %v to %v, meaning there is a hole.\n",
						run.Solid.Triangles[triangleIndex].Vertices[(e+1)%3],
						run.Solid.Triangles[triangleIndex].Vertices[e])
				}
			}
		}
	}
	if len(triangleErrorsMap) != 0 {
		fmt.Fprintf(out, "Errors found in %d/%d triangles\n", len(triangleErrorsMap), len(run.Solid.Triangles))
	}
}

func (run *Run) License() {
	fmt.Fprintln(os.Stderr,
		`The MIT License (MIT)

Copyright (c) 2014 Hagen Schendel

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
`)
}

func (run *Run) Fixnorm() {
	run.Solid.RecalculateNormals()
}

func (run *Run) SetCommand() {
	switch run.Params.Command {
	case "help":
		run.Command = (*Run).Help
		run.DoReadInput = false
	case "copy":
		run.Command = (*Run).Copy
	case "measure":
		run.Command = (*Run).Measure
	case "scale":
		run.Command = (*Run).Scale
	case "translate":
		run.Command = (*Run).Translate
	case "rotate":
		run.Command = (*Run).Rotate
	case "fitbox":
		run.Command = (*Run).FitBox
	case "transform":
		run.Command = (*Run).Transform
	case "license":
		run.Command = (*Run).License
		run.DoReadInput = false
	case "validate":
		run.Command = (*Run).Validate
	default:
		run.Command = (*Run).Help
		run.DoReadInput = false
		os.Stderr.WriteString("Error: unknown command\n")
		run.ExitCode = 11
	}
}

func (run *Run) ExecCommand() {
	run.Command(run)
}

func main() {
	var run Run
	run.ParseParams()
	run.DoReadInput = true
	run.SetCommand()
	if run.DoReadInput {
		if !run.ReadInput() {
			os.Exit(run.ExitCode)
		}
	}
	run.ExecCommand()
	if run.DoWriteOutput {
		if run.Params.Fixnorm {
			run.Fixnorm()
		}
		if !run.WriteOutput() {
			os.Exit(run.ExitCode)
		}
	}
	os.Exit(run.ExitCode)
}
