package core

import "testing"

func TestParseLine(t *testing.T){
	res,err:= ParseLine("* * * * * echo hello")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res[0],res[1])
}
