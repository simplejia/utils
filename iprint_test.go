package utils

import "time"

type P struct {
	I int
}

type C struct {
	P
	J string
	T time.Time
	M map[int]*P
}

func ExampleIprintD() {
	var v interface{}
	IprintD(v)
	v = 1
	IprintD(v)
	v = "str"
	IprintD(v)
	v = &P{I: 1}
	IprintD(v)
	v = &C{P: P{I: 1}, J: "a", M: map[int]*P{1: &P{I: 2}}}
	IprintD(v)
	v = map[string]int{"a": 1}
	IprintD(v)
	v = map[int]int{1: 2}
	IprintD(v)
	v = map[interface{}]int{1: 2}
	IprintD(v)
	i := 1
	v = map[*int]int{&i: 2}
	IprintD(v)
	v = map[int]*C{1: &C{P: P{I: 1}, J: "a", M: map[int]*P{1: &P{I: 2}}}}
	IprintD(v)

	// Output:
	// null
	// 1
	// "str"
	// {
	//   "I": 1
	// }
	// {
	//   "J": "a",
	//   "M": {
	//     "1": {
	//       "I": 2
	//     }
	//   },
	//   "P": {
	//     "I": 1
	//   },
	//   "T": "0001-01-01T00:00:00Z"
	// }
	// {
	//   "a": 1
	// }
	// {
	//   "1": 2
	// }
	// {
	//   "1": 2
	// }
	// {
	//   "1": 2
	// }
	// {
	//   "1": {
	//     "J": "a",
	//     "M": {
	//       "1": {
	//         "I": 2
	//       }
	//     },
	//     "P": {
	//       "I": 1
	//     },
	//     "T": "0001-01-01T00:00:00Z"
	//   }
	// }
}
