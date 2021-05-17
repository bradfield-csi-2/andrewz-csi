package main

import (
  "testing"
)


func BenchmarkNaive(b *testing.B) {
        // run the Fib function b.N times
        for n := 0; n < b.N; n++ {
                Naive()
        }
}


func BenchmarkWithCond(b *testing.B) {
        // run the Fib function b.N times
        for n := 0; n < b.N; n++ {
                WithCond()
        }
}
