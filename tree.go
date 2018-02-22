package main

import (
  "flag"
  "fmt"
)

type Tree struct {
	Left  *Tree
  Value int
  Right *Tree
}

func _walk(t *Tree, ch chan int) {
  if t != nil {
    _walk(t.Left, ch)
    ch <- t.Value
    _walk(t.Right, ch)
  }
}

// Walk walks the tree t sending all values
// from the tree to the channel ch.
func Walk(t *Tree, ch chan int) {
  _walk(t, ch)
  close(ch)
}

// Same determines whether the trees
// t1 and t2 contain the same values.
func Same(t1, t2 *Tree) bool {
  c1 := make(chan int)
  c2 := make(chan int)
  
  go Walk(t1, c1)
  go Walk(t2, c2)
  
  for {
    v1, ok1 := <-c1
    v2, ok2 := <-c2
    
    if v1 != v2 || ok1 != ok2{
      return false
    }
    
    if !ok1 {
      break
    }
  }
  return true
}

func main() { 
  hWorkers := flag.Int("hash-workers", 1, "number of workers on hashing")
  dWorkers := flag.Int("data-workers", 1, "number of workers on data")
  cWorkers := flag.Int("comp-workers", 1, "number of workers on comparison")
  
  //input := flag.String("input", "", "path to input file")
  var input string
  flag.StringVar(&input, "input", "", "path to input file")

  flag.Parse()

  fmt.Println("Number of Hash Workers: ", *hWorkers)
  fmt.Println("Number of Data Workers: ", *dWorkers)
  fmt.Println("Number of Comparison Workers: ", *cWorkers)
  fmt.Println("Input path: ", input)
}

