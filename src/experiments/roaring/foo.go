package main

import (
	"bytes"
	"fmt"
	"math/rand"

	"github.com/RoaringBitmap/roaring"
)

func main() {
	// example inspired by https://github.com/fzandona/goroar
	fmt.Println("==roaring==")
	rb1 := roaring.BitmapOf(1, 2, 3, 4, 5, 100, 1000, 1002, 1004, 1005)
	for i := 0; i < 10000; i++ {
		rb1.Add(uint32(rand.Intn(100000)))
	}

	b, err := rb1.MarshalBinary()
	if err != nil {
		fmt.Println("err : ", err)
	}
	fmt.Println("len :", len(b), "vs len :", len(rb1.String()))

	rb2 := roaring.BitmapOf(3, 4, 1000)
	fmt.Println(rb2.String())

	rb3 := roaring.NewBitmap()
	fmt.Println(rb3.String())

	fmt.Println("Cardinality: ", rb1.GetCardinality())

	fmt.Println("Contains 3? ", rb1.Contains(3))

	rb1.And(rb2)
	rb3.Add(1)
	rb3.Add(5)

	rb3.Or(rb1)

	// prints 1, 3, 4, 5, 1000
	i := rb3.Iterator()
	for i.HasNext() {
		fmt.Println(i.Next())
	}
	fmt.Println()

	// next we include an example of serialization
	buf := new(bytes.Buffer)
	rb1.WriteTo(buf) // we omit error handling
	newrb := roaring.NewBitmap()
	newrb.ReadFrom(buf)
	if rb1.Equals(newrb) {
		fmt.Println("I wrote the content to a byte stream and read it back.")
	}
}
