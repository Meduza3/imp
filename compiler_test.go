package main_test

import (
	"context"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"testing"
	"time"
)

func testAssembly(t *testing.T, inputCode string, expectedOutputNumbers []int, userInput string) {
	t.Helper()
	// TODO: implement the equivalent logic:
	// 1) Write inputCode to a file
	file, err := os.Create("test_code.imp")
	if err != nil {
		t.Fatalf("failed to create file")
	}
	if _, err := file.WriteString(inputCode); err != nil {
		t.Fatalf("failed to write input code: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("failed to close file: %v", err)
	}
	// 2) Run compiler
	makeCmd := exec.Command("make")
	if err := makeCmd.Run(); err != nil {
		t.Fatalf("failed to execute make: %v", err)
	}

	file2, err := os.Create("test_code.mr")
	if err != nil {
		t.Fatalf("failed to create file")
	}
	compilerCmd := exec.Command("./bin/main", file.Name(), file2.Name())

	if err := compilerCmd.Start(); err != nil {
		t.Fatalf("failed to execute compiler: %v", err)
	}
	compilerCmd.Wait()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// 3) Run VM / big-numbers VM
	cmd2 := exec.CommandContext(ctx, "resources/maszyna_wirtualna/maszyna-wirtualna", "test_code.mr")
	vmPipe, _ := cmd2.StdoutPipe()
	vmInput, _ := cmd2.StdinPipe()
	// 4) Capture & compare the outputs
	if err := cmd2.Start(); err != nil {
		t.Fatalf("failed to execute vm: %v", err)
	}
	if userInput != "" {
		// e.g. "10\n20\n" if the program needs two integers.
		if _, err := io.WriteString(vmInput, userInput); err != nil {
			t.Fatalf("failed to write to VM stdin: %v", err)
		}
	}
	output, err := io.ReadAll(vmPipe)
	if err != nil {
		t.Fatalf("failed to read compiler output: %v", err)
	}
	if err := cmd2.Wait(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			t.Fatalf("VM execution timed out after 20s")
		}
		t.Fatalf("VM process failed: %v", err)
	}
	nums := extractNumbers(string(output))
	// 5) Validate against expectedOutputNumbers
	if len(nums) != len(expectedOutputNumbers) {

		t.Fatalf("expected %d numbers, but got %d: %v", len(expectedOutputNumbers), len(nums), nums)
	}

	for i, expected := range expectedOutputNumbers {
		if nums[i] != expected {
			t.Fatalf("expected output %d at index %d, but got %d", expected, i, nums[i])
		}
	}
}

func extractNumbers(output string) []int {
	// Define regex pattern to match numbers after ">"
	re := regexp.MustCompile(`>\s*(-?\d+)`)

	// Find all matches
	matches := re.FindAllStringSubmatch(output, -1)

	// Convert matches to integers
	var numbers []int
	for _, match := range matches {
		if len(match) > 1 { // Ensure there is a captured group
			num, err := strconv.Atoi(match[1])
			if err == nil {
				numbers = append(numbers, num)
			}
		}
	}

	return numbers
}

func TestIfStatements(t *testing.T) {
	// t.Skip()
	cases := []struct {
		inputCode      string
		expectedOutput []int
	}{
		{"PROGRAM IS BEGIN IF 15 = 17 THEN WRITE 1; ENDIF END", []int{}},
		{"PROGRAM IS BEGIN IF 17 = 17 THEN WRITE 1; ENDIF END", []int{1}},
		{"PROGRAM IS BEGIN IF 15 = 17 THEN WRITE 1; ELSE WRITE 0; ENDIF END", []int{0}},
		{"PROGRAM IS BEGIN IF 17 = 17 THEN WRITE 1; ELSE WRITE 0; ENDIF END", []int{1}},
		{"PROGRAM IS BEGIN IF 15 != 17 THEN WRITE 1; ENDIF END", []int{1}},
		{"PROGRAM IS BEGIN IF 17 != 17 THEN WRITE 1; ENDIF END", []int{}},
		{"PROGRAM IS BEGIN IF 15 != 17 THEN WRITE 1; ELSE WRITE 0; ENDIF END", []int{1}},
		{"PROGRAM IS BEGIN IF 17 != 17 THEN WRITE 1; ELSE WRITE 0; ENDIF END", []int{0}},
		{"PROGRAM IS BEGIN IF 15 > 17 THEN WRITE 1; ENDIF END", []int{}},
		{"PROGRAM IS BEGIN IF 17 > 17 THEN WRITE 1; ENDIF END", []int{}},
		{"PROGRAM IS BEGIN IF 17 > 15 THEN WRITE 1; ENDIF END", []int{1}},
		{"PROGRAM IS BEGIN IF 15 > 17 THEN WRITE 1; ELSE WRITE 0; ENDIF END", []int{0}},
		{"PROGRAM IS BEGIN IF 17 > 17 THEN WRITE 1; ELSE WRITE 0; ENDIF END", []int{0}},
		{"PROGRAM IS BEGIN IF 17 > 15 THEN WRITE 1; ELSE WRITE 0; ENDIF END", []int{1}},
		{"PROGRAM IS BEGIN IF 15 < 17 THEN WRITE 1; ENDIF END", []int{1}},
		{"PROGRAM IS BEGIN IF 17 < 17 THEN WRITE 1; ENDIF END", []int{}},
		{"PROGRAM IS BEGIN IF 17 < 15 THEN WRITE 1; ENDIF END", []int{}},
		{"PROGRAM IS BEGIN IF 15 < 17 THEN WRITE 1; ELSE WRITE 0; ENDIF END", []int{1}},
		{"PROGRAM IS BEGIN IF 17 < 17 THEN WRITE 1; ELSE WRITE 0; ENDIF END", []int{0}},
		{"PROGRAM IS BEGIN IF 17 < 15 THEN WRITE 1; ELSE WRITE 0; ENDIF END", []int{0}},
		{"PROGRAM IS BEGIN IF 15 >= 17 THEN WRITE 1; ENDIF END", []int{}},
		{"PROGRAM IS BEGIN IF 17 >= 17 THEN WRITE 1; ENDIF END", []int{1}},
		{"PROGRAM IS BEGIN IF 17 >= 15 THEN WRITE 1; ENDIF END", []int{1}},
		{"PROGRAM IS BEGIN IF 15 >= 17 THEN WRITE 1; ELSE WRITE 0; ENDIF END", []int{0}},
		{"PROGRAM IS BEGIN IF 17 >= 17 THEN WRITE 1; ELSE WRITE 0; ENDIF END", []int{1}},
		{"PROGRAM IS BEGIN IF 17 >= 15 THEN WRITE 1; ELSE WRITE 0; ENDIF END", []int{1}},
		{"PROGRAM IS BEGIN IF 15 <= 17 THEN WRITE 1; ENDIF END", []int{1}},
		{"PROGRAM IS BEGIN IF 17 <= 17 THEN WRITE 1; ENDIF END", []int{1}},
		{"PROGRAM IS BEGIN IF 17 <= 15 THEN WRITE 1; ENDIF END", []int{}},
		{"PROGRAM IS BEGIN IF 15 <= 17 THEN WRITE 1; ELSE WRITE 0; ENDIF END", []int{1}},
		{"PROGRAM IS BEGIN IF 17 <= 17 THEN WRITE 1; ELSE WRITE 0; ENDIF END", []int{1}},
		{"PROGRAM IS BEGIN IF 17 <= 15 THEN WRITE 1; ELSE WRITE 0; ENDIF END", []int{0}},
	}

	for _, tt := range cases {
		t.Run(tt.inputCode, func(t *testing.T) {
			testAssembly(t, tt.inputCode, tt.expectedOutput, "")
		})
	}
}

func TestWhileLoop(t *testing.T) {
	inputCode := "PROGRAM IS i BEGIN i := 1; WHILE i <= 10 DO WRITE i; i := i + 1; ENDWHILE END"
	expectedOutput := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	testAssembly(t, inputCode, expectedOutput, "")
}

func TestRepeatLoop(t *testing.T) {
	inputCode := "PROGRAM IS i BEGIN i := 1; REPEAT WRITE i; i := i + 1; UNTIL i > 10; END"
	expectedOutput := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	testAssembly(t, inputCode, expectedOutput, "")
}

func TestForLoops(t *testing.T) {
	cases := []struct {
		inputCode      string
		expectedOutput []int
	}{
		{"PROGRAM IS BEGIN FOR i FROM 5 TO 10 DO WRITE i; ENDFOR END", []int{5, 6, 7, 8, 9, 10}},
		{"PROGRAM IS BEGIN FOR i FROM 10 DOWNTO 5 DO WRITE i; ENDFOR END", []int{10, 9, 8, 7, 6, 5}},
	}

	for _, tt := range cases {
		t.Run(tt.inputCode, func(t *testing.T) {
			testAssembly(t, tt.inputCode, tt.expectedOutput, "")
		})
	}
}

func TestReadStatements(t *testing.T) {
	cases := []struct {
		inputCode      string
		expectedOutput []int
		userInput      string
	}{
		{"PROGRAM IS x BEGIN READ x; WRITE x; END", []int{123}, "123\n"},
		{"PROGRAM IS x[-5:5] BEGIN READ x[-2]; WRITE x[-2]; END", []int{124}, "124\n"},
		{"PROGRAM IS x[-5:5] BEGIN READ x[2]; WRITE x[2]; END", []int{125}, "125\n"},
		{"PROGRAM IS x[-5:5], n BEGIN n := -5; READ x[n]; WRITE x[n]; END", []int{126}, "126\n"},
		{"PROGRAM IS n, x[-5:5] BEGIN n := 5; READ x[n]; WRITE x[n]; END", []int{127}, "127\n"},
	}

	for _, tt := range cases {
		t.Run(tt.inputCode, func(t *testing.T) {
			testAssembly(t, tt.inputCode, tt.expectedOutput, tt.userInput)
		})
	}
}

func TestAssignmentStatements(t *testing.T) {
	cases := []struct {
		inputCode      string
		expectedOutput []int
	}{
		{"PROGRAM IS x BEGIN x := 123; WRITE x; END", []int{123}},
		{"PROGRAM IS x[-10:-5] BEGIN x[-7] := 124; WRITE x[-7]; END", []int{124}},
		{"PROGRAM IS x[5:10] BEGIN x[7] := 125; WRITE x[7]; END", []int{125}},
		{"PROGRAM IS x[-5:5] BEGIN x[0] := 126; WRITE x[0]; END", []int{126}},
		{"PROGRAM IS n, x[5:10] BEGIN n := 7; x[n] := 127; WRITE x[n]; END", []int{127}},
		{"PROGRAM IS x[5:10], n BEGIN n := 7; x[n] := 128; WRITE x[n]; END", []int{128}},
	}

	for _, tt := range cases {
		t.Run(tt.inputCode, func(t *testing.T) {
			testAssembly(t, tt.inputCode, tt.expectedOutput, "")
		})
	}
}

func TestArithmeticOperations(t *testing.T) {
	cases := []struct {
		inputCode      string
		expectedOutput []int
	}{
		{"PROGRAM IS x BEGIN x := 15 + 17; WRITE x; END", []int{32}},
		{"PROGRAM IS x BEGIN x := 15 - 17; WRITE x; END", []int{-2}},
		{"PROGRAM IS x, y, z BEGIN y := 3; z := 4; x := y + z; WRITE x; END", []int{7}},
		{"PROGRAM IS x, y, z BEGIN y := 3; z := 4; x := y - z; WRITE x; END", []int{-1}},
		{"PROGRAM IS x, y[1:10], z[1:10], n BEGIN n := 5; y[n] := 3; z[n] := 4; x := y[n] + z[n]; WRITE x; END", []int{7}},
		{"PROGRAM IS x, y BEGIN y := 15; x := y + 17; WRITE x; END", []int{32}},
		{"PROGRAM IS x, y BEGIN y := 17; x := 15 + y; WRITE x; END", []int{32}},
		{"PROGRAM IS x, y BEGIN y := 15; x := y - 17; WRITE x; END", []int{-2}},
		{"PROGRAM IS x, y BEGIN y := 17; x := 15 - y; WRITE x; END", []int{-2}},
		{"PROGRAM IS x, y[5:10], n BEGIN n := 5; y[n] := 17; x := 15 + y[n]; WRITE x; END", []int{32}},
		{"PROGRAM IS x BEGIN x := 24 * 11; WRITE x; END", []int{264}},
		{"PROGRAM IS x BEGIN x := 16 * 8; WRITE x; END", []int{128}},
		{"PROGRAM IS x BEGIN x := 7 * 5; WRITE x; END", []int{35}},
		{"PROGRAM IS x BEGIN x := 5 * 7; WRITE x; END", []int{35}},
		{"PROGRAM IS x BEGIN x := -5 * 7; WRITE x; END", []int{-35}},
		{"PROGRAM IS x BEGIN x := -5 * -7; WRITE x; END", []int{35}},
		{"PROGRAM IS x BEGIN x := 5 * -7; WRITE x; END", []int{-35}},
		{"PROGRAM IS x, y, z BEGIN y := 5; z := 7; x := y * z; WRITE x; END", []int{35}},
		{"PROGRAM IS x BEGIN x := 12 / 2; WRITE x; END", []int{6}},
		{"PROGRAM IS x BEGIN x := 17 / 2; WRITE x; END", []int{8}},
		{"PROGRAM IS x BEGIN x := 9 / -2; WRITE x; END", []int{-5}},
		{"PROGRAM IS x BEGIN x := 9 / 2; WRITE x; END", []int{4}},
		{"PROGRAM IS x BEGIN x := 10 / -2; WRITE x; END", []int{-5}},
		{"PROGRAM IS x BEGIN x := 10 / 2; WRITE x; END", []int{5}},
		{"PROGRAM IS x BEGIN x := 10 / 0; WRITE x; END", []int{0}},
		{"PROGRAM IS x, t[1:10] BEGIN t[5] := 17; x := t[5] / 2; WRITE x; END", []int{8}},
		{"PROGRAM IS x BEGIN x := 12 % 2; WRITE x; END", []int{0}},
		{"PROGRAM IS x BEGIN x := 13 % 2; WRITE x; END", []int{1}},
		{"PROGRAM IS x BEGIN x := 127 % 12; WRITE x; END", []int{7}},
		{"PROGRAM IS x BEGIN x := 10 % 3; WRITE x; END", []int{1}},
		{"PROGRAM IS x BEGIN x := 10 % -3; WRITE x; END", []int{-2}},
		{"PROGRAM IS x BEGIN x := -10 % 3; WRITE x; END", []int{2}},
		{"PROGRAM IS x BEGIN x := -10 % -3; WRITE x; END", []int{-1}},
		{"PROGRAM IS x BEGIN x := 8968 % -8; WRITE x; END", []int{0}},
		{"PROGRAM IS x BEGIN x := -5467 % 11; WRITE x; END", []int{0}},
		{"PROGRAM IS x BEGIN x := 548 % -2901; WRITE x; END", []int{-2353}},
	}

	for _, tt := range cases {
		t.Run(tt.inputCode, func(t *testing.T) {
			testAssembly(t, tt.inputCode, tt.expectedOutput, "")
		})
	}
}

func TestWrite(t *testing.T) {
	cases := []struct {
		input_code     string
		expectedOutput []int
		userInput      string
	}{
		{
			"PROGRAM IS BEGIN WRITE 123; WRITE 124; END",
			[]int{123, 124},
			"",
		},
		{
			"PROGRAM IS x BEGIN READ x; WRITE x; END",
			[]int{123},
			"123\n",
		},
		{
			"PROGRAM IS x[-5:5] BEGIN READ x[-2]; WRITE x[-2]; END",
			[]int{124},
			"124\n",
		},
		{
			"PROGRAM IS x[-5:5] BEGIN READ x[2]; WRITE x[2]; END",
			[]int{125},
			"125\n",
		},
		{
			"PROGRAM IS x[-5:5], n BEGIN n := -5; READ x[n]; WRITE x[n]; END",
			[]int{126},
			"126\n",
		},
		{
			"PROGRAM IS n, x[-5:5] BEGIN n := 5; READ x[n]; WRITE x[n]; END",
			[]int{127},
			"127\n",
		},
	}

	for _, tt := range cases {
		t.Run(tt.input_code, func(t *testing.T) {
			testAssembly(t, tt.input_code, tt.expectedOutput, tt.userInput)
		})
	}
}

func TestProc(t *testing.T) {
	cases := []struct {
		input_code     string
		expectedOutput []int
		userInput      string
	}{
		{
			"PROCEDURE write_proc(proc_x) IS BEGIN WRITE proc_x; END PROGRAM IS main_x BEGIN main_x := 123; write_proc(main_x); END",
			[]int{123},
			"",
		},
		{
			"PROCEDURE read_proc(proc_x) IS BEGIN READ proc_x; END PROGRAM IS main_x BEGIN read_proc(main_x); WRITE main_x; END",
			[]int{124},
			"124\n",
		},
		{
			"PROCEDURE add(x, y, z) IS BEGIN z := x + y; END PROGRAM IS x, y, z BEGIN x := 15; y := 17; add(x, y, z); WRITE x; WRITE y; WRITE z; END",
			[]int{15, 17, 32},
			"",
		},
		{
			"PROCEDURE add(T x, T y, T z) IS BEGIN z[3] := x[1] + y[2]; END PROGRAM IS x[-5:5], y[0:5], z[-10:10] BEGIN x[1] := 15; y[2] := 17; add(x, y, z); WRITE x[1]; WRITE y[2]; WRITE z[3]; END",
			[]int{15, 17, 32},
			"",
		},
		{
			"PROCEDURE add(T x, T y, T z) IS a, b, c BEGIN a := 3; b := 1; c := 2; z[a] := x[b] + y[c]; END PROGRAM IS x[-5:5], y[0:5], z[-10:10] BEGIN x[1] := 15; y[2] := 17; add(x, y, z); WRITE x[1]; WRITE y[2]; WRITE z[3]; END",
			[]int{15, 17, 32},
			"",
		},
		{
			"PROCEDURE add(T x, T y, T z, a, b, c) IS BEGIN z[a] := x[b] + y[c]; END PROGRAM IS x[-5:5], y[0:5], z[-10:10], a, b, c BEGIN x[1] := 15; y[2] := 17; a := 3; b := 1; c := 2; add(x, y, z, a, b, c); WRITE x[1]; WRITE y[2]; WRITE z[3]; END",
			[]int{15, 17, 32},
			"",
		},
		{
			"PROCEDURE modify(T t) IS BEGIN t[10] := 125; END PROCEDURE modify_in_another_proc(T t) IS BEGIN modify(t); END PROGRAM IS t[-10:10] BEGIN modify_in_another_proc(t); WRITE t[10]; END",
			[]int{125},
			"",
		},
		{
			"PROCEDURE modify(T t, n) IS BEGIN t[n] := 126; END PROCEDURE modify_in_another_proc(T t, n) IS BEGIN modify(t, n); END PROGRAM IS t[-10:10], n BEGIN n := 10; modify_in_another_proc(t, n); WRITE t[10]; END",
			[]int{126},
			"",
		},
		{
			"PROCEDURE proc(T t) IS x BEGIN x := t[10]; WRITE x; END PROGRAM IS t[-10:10] BEGIN t[10] := 123; proc(t); END",
			[]int{123},
			"",
		},
		{
			"PROCEDURE proc(T t, n) IS x BEGIN x := t[n]; WRITE x; END PROGRAM IS t[-10:10], n BEGIN n := 10; t[10] := 123; proc(t, n); END",
			[]int{123},
			"",
		},
		{
			"PROCEDURE proc(T t) IS x BEGIN x := t[10] + t[11]; WRITE x; END PROGRAM IS t[-10:15] BEGIN t[10] := 123; t[11] := 2; proc(t); END",
			[]int{125},
			"",
		},
	}

	for _, tt := range cases {
		t.Run(tt.input_code, func(t *testing.T) {
			testAssembly(t, tt.input_code, tt.expectedOutput, tt.userInput)
		})
	}
}

func TestLast(t *testing.T) {
	cases := []struct {
		inputCode      string
		expectedOutput []int
		userInput      string
	}{
		{
			`
PROCEDURE write(T t, n) IS
BEGIN
WRITE n;
END

PROGRAM IS
  t[1:23], n
BEGIN
  n:=23;
  write(t,n);
  WRITE 1234567890;
END
`,
			[]int{23, 1234567890},
			"",
		},
	}
	for _, tt := range cases {
		t.Run(tt.inputCode, func(t *testing.T) {
			testAssembly(t, tt.inputCode, tt.expectedOutput, tt.userInput)
		})
	}
}

// TestBCProgram tests the "bc" program that calculates a binomial coefficient.
// For example, with inputs n=20 and k=9 the expected output is 167960.
func TestBCProgram(t *testing.T) {
	inputCode := `
PROCEDURE factorial(T s,n) IS
  p
BEGIN
  s[0]:=1;
  p:=s[0];
  FOR i FROM 1 TO n DO
    s[i]:=p*i;
    p:=s[i];
  ENDFOR
END

PROCEDURE bc(n,k,m) IS
  s[0:100],p
BEGIN
  factorial(s,n);
  p:=n-k;
  m:=s[n]/s[k];
  m:=m/s[p];
END

PROGRAM IS
    n,k,w
BEGIN
    READ n;
    READ k;
    bc(n,k,w);
    WRITE w;
END
`
	// The test uses n=20 and k=9. Expected result: 167960.
	expected := []int{167960}
	userInput := "20\n9\n"
	testAssembly(t, inputCode, expected, userInput)
}

// TestArrayManipulationProgram tests the integrated program that creates three arrays
// and fills tc with values computed from arrays ta and tb.
// The program is expected to write 25 numbers corresponding to indices 0..24.
func TestArrayManipulationProgram(t *testing.T) {
	inputCode := `
PROGRAM IS
    n, j, ta[0:24], tb[0:24], tc[0:24]
BEGIN
    n:=24;
    tc[0]:=n;
    tc[n]:=n-tc[0];
    j:=tc[0]+1;
    FOR i FROM tc[0] DOWNTO tc[n] DO
        ta[i]:=i+1;
        tb[i]:=j-i;
    ENDFOR
    j:=tc[n];
    WHILE j<tc[0] DO
        tc[j]:=ta[j]*tb[j];
        j:=j+1;
    ENDWHILE
    FOR i FROM 0 TO n DO
        WRITE tc[i];
    ENDFOR
END
`
	// Let's simulate the program:
	//   n=24, tc[0]=24, tc[24]=24-24=0, j initially = 25.
	//   FOR i FROM 24 DOWNTO 0: ta[i] = i+1, tb[i] = 25 - i.
	//   Then j is set to tc[24] (0) and WHILE j < 24:
	//       tc[j] = ta[j] * tb[j] = (j+1) * (25-j) for j=0,...,23.
	//   tc[24] remains 0.
	// The expected outputs are:
	//   tc[0] = 1*25 = 25,
	//   tc[1] = 2*24 = 48,
	//   tc[2] = 3*23 = 69,
	//   tc[3] = 4*22 = 88,
	//   tc[4] = 5*21 = 105,
	//   tc[5] = 6*20 = 120,
	//   tc[6] = 7*19 = 133,
	//   tc[7] = 8*18 = 144,
	//   tc[8] = 9*17 = 153,
	//   tc[9] = 10*16 = 160,
	//   tc[10]= 11*15 = 165,
	//   tc[11]= 12*14 = 168,
	//   tc[12]= 13*13 = 169,
	//   tc[13]= 14*12 = 168,
	//   tc[14]= 15*11 = 165,
	//   tc[15]= 16*10 = 160,
	//   tc[16]= 17*9  = 153,
	//   tc[17]= 18*8  = 144,
	//   tc[18]= 19*7  = 133,
	//   tc[19]= 20*6  = 120,
	//   tc[20]= 21*5  = 105,
	//   tc[21]= 22*4  = 88,
	//   tc[22]= 23*3  = 69,
	//   tc[23]= 24*2  = 48,
	//   tc[24]= 0.
	expected := []int{
		25, 48, 69, 88, 105, 120, 133, 144, 153, 160,
		165, 168, 169, 168, 165, 160, 153, 144, 133, 120,
		105, 88, 69, 48, 0,
	}
	testAssembly(t, inputCode, expected, "")
}
