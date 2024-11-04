package server

import "glover/components"

// var locations = map[int]Location{
// 	// Left, 1st row
// 	0: {0, 0},
// 	1: {0, 1},
// 	2: {0, 2},
// 	3: {0, 3},
// 	4: {0, 4},
//
// 	// Right, first row
// 	5: {0, 2},
// 	6: {0, 3},
// 	7: {0, 4},
// 	8: {0, 5},
// 	9: {0, 6},
//
// 	// Left, second row
// 	10: {1, 0},
// 	11: {1, 1},
// 	12: {1, 2},
// 	13: {1, 3},
// 	14: {1, 4},
// 	15: {1, 5},
//
// 	// Right, second row
// 	16: {1, 1},
// 	17: {1, 2},
// 	18: {1, 3},
// 	19: {1, 4},
// 	20: {1, 5},
// 	21: {1, 6},
//
// 	// Left, third row
// 	22: {2, 0},
// 	23: {2, 1},
// 	24: {2, 2},
// 	25: {2, 3},
// 	26: {2, 4},
// 	27: {2, 5},
//
// 	// Right, third row
// 	28: {2, 1},
// 	29: {2, 2},
// 	30: {2, 3},
// 	31: {2, 4},
// 	32: {2, 5},
// 	33: {2, 6},
//
// 	// Left, fourth row
// 	34: {3, 0},
// 	35: {3, 1},
// 	36: {3, 2},
// 	37: {3, 3},
// 	38: {3, 4},
// 	39: {3, 5},
//
// 	// Right, fourth row
// 	40: {3, 1},
// 	41: {3, 2},
// 	42: {3, 3},
// 	43: {3, 4},
// 	44: {3, 5},
// 	45: {3, 6},
//
// 	// Left, fifth row
// 	46: {4, 0},
// 	47: {4, 1},
// 	48: {4, 2},
// 	49: {4, 3},
// 	50: {4, 4},
// 	51: {4, 5},
//
// 	52: {0, 6},
// 	53: {1, 6},
// 	54: {2, 6},
//
// 	// Right, fifth row
// 	55: {2, 0},
// 	56: {1, 0},
// 	57: {0, 0},
//
// 	58: {4, 1},
// 	59: {4, 2},
// 	60: {4, 3},
// 	61: {4, 4},
// 	62: {4, 5},
// 	63: {4, 6},
//
// 	// Left, last row
// 	64: {5, 0},
// 	65: {5, 1},
// 	66: {5, 2},
// 	67: {5, 3},
// 	68: {5, 4},
//
// 	70: {3, 6},
// 	71: {4, 6},
// 	72: {5, 6},
//
// 	// Right, fifth row
// 	73: {5, 0},
// 	74: {4, 0},
// 	75: {3, 0},
//
// 	76: {5, 2},
// 	77: {5, 3},
// 	78: {5, 4},
// 	79: {5, 5},
// }

// TODO: This is super-tied to Glove80. I don't know how to automate this because
// position assignment seems sort of random in regards to target position, and
// other keyboards are properly the same. Maybe make this configurable?
// info.json for zmk-configurator seems close enough...
// https://github.com/nickcoutsos/keymap-editor-contrib
var locationsOnGrid = map[int]components.Location{
	// Left, 1st row
	0: {Row: 0, Col: 0},
	1: {Row: 0, Col: 1},
	2: {Row: 0, Col: 2},
	3: {Row: 0, Col: 3},
	4: {Row: 0, Col: 4},

	// Right,Col: first row
	5: {Row: 0, Col: 7},
	6: {Row: 0, Col: 8},
	7: {Row: 0, Col: 9},
	8: {Row: 0, Col: 10},
	9: {Row: 0, Col: 11},

	// Left,Col: second row
	10: {Row: 1, Col: 0},
	11: {Row: 1, Col: 1},
	12: {Row: 1, Col: 2},
	13: {Row: 1, Col: 3},
	14: {Row: 1, Col: 4},
	15: {Row: 1, Col: 5},

	// Right,Col: second row
	16: {Row: 1, Col: 6},
	17: {Row: 1, Col: 7},
	18: {Row: 1, Col: 8},
	19: {Row: 1, Col: 9},
	20: {Row: 1, Col: 10},
	21: {Row: 1, Col: 11},

	// Left,Col: third row
	22: {Row: 2, Col: 0},
	23: {Row: 2, Col: 1},
	24: {Row: 2, Col: 2},
	25: {Row: 2, Col: 3},
	26: {Row: 2, Col: 4},
	27: {Row: 2, Col: 5},

	// Right,Col: third row
	28: {Row: 2, Col: 6},
	29: {Row: 2, Col: 7},
	30: {Row: 2, Col: 8},
	31: {Row: 2, Col: 9},
	32: {Row: 2, Col: 10},
	33: {Row: 2, Col: 11},

	// Left,Col: fourth row
	34: {Row: 3, Col: 0},
	35: {Row: 3, Col: 1},
	36: {Row: 3, Col: 2},
	37: {Row: 3, Col: 3},
	38: {Row: 3, Col: 4},
	39: {Row: 3, Col: 5},

	// Right,Col: fourth row
	40: {Row: 3, Col: 6},
	41: {Row: 3, Col: 7},
	42: {Row: 3, Col: 8},
	43: {Row: 3, Col: 9},
	44: {Row: 3, Col: 10},
	45: {Row: 3, Col: 11},

	// Left,Col: fifth row
	46: {Row: 4, Col: 0},
	47: {Row: 4, Col: 1},
	48: {Row: 4, Col: 2},
	49: {Row: 4, Col: 3},
	50: {Row: 4, Col: 4},
	51: {Row: 4, Col: 5},

	// Left cluster,Col: first row
	52: {Row: 6, Col: 2},
	53: {Row: 6, Col: 3},
	54: {Row: 6, Col: 4},

	// Right cluster,Col: first row
	55: {Row: 6, Col: 7},
	56: {Row: 6, Col: 8},
	57: {Row: 6, Col: 9},

	// Right,Col: fifth row
	58: {Row: 4, Col: 6},
	59: {Row: 4, Col: 7},
	60: {Row: 4, Col: 8},
	61: {Row: 4, Col: 9},
	62: {Row: 4, Col: 10},
	63: {Row: 4, Col: 11},

	// Left,Col: last row
	64: {Row: 5, Col: 0},
	65: {Row: 5, Col: 1},
	66: {Row: 5, Col: 2},
	67: {Row: 5, Col: 3},
	68: {Row: 5, Col: 4},

	// Left cluster,Col: last row
	69: {Row: 7, Col: 2},
	70: {Row: 7, Col: 3},
	71: {Row: 7, Col: 4},

	// Right cluster,Col: last row
	72: {Row: 7, Col: 7},
	73: {Row: 7, Col: 8},
	74: {Row: 7, Col: 9},

	// Right. last row
	75: {Row: 5, Col: 7},
	76: {Row: 5, Col: 8},
	77: {Row: 5, Col: 9},
	78: {Row: 5, Col: 10},
	79: {Row: 5, Col: 11},
}
