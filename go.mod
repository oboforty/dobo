module github.com/oboforty/dobo

go 1.23.4

require (
	github.com/bits-and-blooms/bloom/v3 v3.7.0
	github.com/emirpasic/gods/v2 v2.0.0-alpha
	github.com/pelletier/go-toml/v2 v2.2.3
)

require github.com/bits-and-blooms/bitset v1.20.0 // indirect

replace github.com/oboforty/dobo/lsm => ../lsm
