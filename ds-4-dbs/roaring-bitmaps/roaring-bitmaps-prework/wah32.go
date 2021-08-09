package bitmap

func deWah32(cData []uint64) []uint64 {
	cLen := cData[0]
	if cLen < 10 {
		return cData[1:]
	}

	uLen := int(cLen)
	uData := make([]uint64, uLen)
	i, j, jLen := 0, 2, len(cData)

	takeLeft := true
	var staged32, staged64, bit31, bit30, fillCntMask, fill0, fill1, rMask uint64
	left, right, space := cData[1], cData[1], uint64(64)
	rMask = (1 << 32) - 1
	left >>= 32
	right &= rMask
	bit31 = 1 << 31
	bit30 = 1 << 30
	staged64 = 0
	fillCntMask = rMask >> 2
	fill0 = 0
	fill1 = (rMask << 32) | rMask

	for i < uLen {
		if takeLeft {
			staged32 = left
		} else if j < jLen {
			staged32 = right
			left, right = cData[j], cData[j]
			left >>= 32
			right &= rMask
			j++
		} else {
			staged32 = right
		}
		takeLeft = !takeLeft

		cType := Literal

		if staged32&bit31 != 0 {
			if staged32&bit30 == 0 {
				cType = Fill0
			} else {
				cType = Fill1
			}
		}

		var cnt, fill uint64
		caseNum := 0
		switch cType {
		case Literal:
			cnt = 31
			caseNum = 1
		case Fill0:
			fillCnt := staged32 & fillCntMask
			cnt = fillCnt * 31
			if cnt < 64+space {
				caseNum = 1
				staged32 = fill0 >> (64 - space)
			} else {
				caseNum = 2
				fill = fill0
			}
		case Fill1:
			fillCnt := staged32 & fillCntMask
			cnt = fillCnt * 31
			if cnt < 64+space {
				staged32 = fill1 >> (64 - space)
				caseNum = 1
			} else {
				caseNum = 2
				fill = fill1
			}

		default:
			panic("cType value not handled")
		}

		switch caseNum {
		case 1:
			if space > cnt {
				staged64 |= staged32 << (space - cnt)
				space -= cnt
			} else {
				staged64 |= staged32 >> (cnt - space)
				uData[i] = staged64
				i++
				space += 64 - cnt //33 //64 - 31
				staged64 = staged32 << space
			}
		case 2:
			finish := fill >> (64 - space)
			cnt -= space
			nFill := cnt / 64
			space = 64 - (cnt % 64)
			ext := fill << space
			staged64 |= finish
			uData[i] = staged64
			i++
			for k := uint64(0); k < nFill; k++ {
				uData[i] = fill
				i++
			}
			staged64 = ext
		default:
			panic("case not handled")
		}

	}

	return uData

}

func wah32(data []uint64) []uint64 {
	var compressed []uint64

	dLen := len(data)

	if dLen < 10 {
		compressed = make([]uint64, dLen+1)
		compressed[0] = uint64(dLen)
		copy(compressed[1:], data)
		return compressed
	}

	split32 := splitWah32(data)
	compress32 := compressWah32(split32)

	compressed = convert32To64(compress32, dLen)
	return compressed
}

func compressWah32(data []uint32) []uint32 {
	var compress32 []uint32

	var staged, fill1, firstFill0, firstFill1, rleMax0, rleMax1, filler131 uint32
	fill1 = (1 << 31) - 1 //uint32(-1)
	fill1 = (fill1 << 1) + 1
	firstFill0 = (1 << 31) + 1
	firstFill1 = (3 << 30) + 1
	rleMax0 = firstFill0 | (fill1 >> 2)
	rleMax1 = fill1
	filler131 = fill1 >> 1

	lastType := Literal
	for _, c := range data {
		cType := Literal
		firstFill := firstFill0
		rleMax := rleMax0
		if c == 0 {
			cType = Fill0
		}
		if c == filler131 {
			cType = Fill1
			firstFill = firstFill1
			rleMax = rleMax1
		}
		swType := (lastType << 2) | cType
		caseNum := 0
		switch swType {
		case LL:
			caseNum = 1
		case LZ:
			caseNum = 2
		case LO:
			caseNum = 2
		case ZL:
			caseNum = 3
		case ZZ:
			caseNum = 5
		case ZO:
			caseNum = 4
		case OL:
			caseNum = 3
		case OZ:
			caseNum = 4
		case OO:
			caseNum = 5
		default:
			panic("swType not recognized")
		}

		switch caseNum {
		case 1:
			compress32 = append(compress32, c)
		case 2:
			staged = firstFill
		case 3:
			compress32 = append(compress32, staged)
			compress32 = append(compress32, c)
		case 4:
			compress32 = append(compress32, staged)
			staged = firstFill
		case 5:
			staged++
			if staged == rleMax {
				compress32 = append(compress32, staged)
				lastType = Literal
				continue
			}
		default:
			panic("can't understand caseNum")
		}
		lastType = cType
	}

	if lastType != Literal {
		compress32 = append(compress32, staged)
	}

	return compress32
}

func convert32To64(data []uint32, origLen int) []uint64 {
	var converted []uint64
	converted = make([]uint64, 1)
	converted[0] = uint64(origLen)

	var staged uint64
	i := 0
	dLen := len(data)
	for i+1 < dLen {
		staged = (uint64(data[i]) << 32) | uint64(data[i+1])
		converted = append(converted, staged)
		i += 2
	}
	if i < dLen {
		staged = (uint64(data[i]) << 32)
		converted = append(converted, staged)
	}
	return converted

}

func splitWah32(data []uint64) []uint32 {

	dLen := len(data)

	if dLen < 10 {
		panic("shouldn't have to handle this case because won't compress")
	}

	i := 2
	var curr, next, staged uint64
	curr, next = data[0], data[1]

	split32 := make([]uint32, 0)
	leftover := 64

	for i < dLen {
		staged = curr << (64 - uint64(leftover))
		leftover -= 31 //int(w - 1)
		if leftover < 0 {
			staged |= next >> uint64(31+leftover)
		}
		if leftover <= 0 {
			leftover += 64
			curr = next
			next = data[i]
			i++
		} //else { //leftover > 0 means there's still some left
		staged >>= 33
		split32 = append(split32, uint32(staged))
	}

	for leftover-31 >= 0 {
		staged = curr << (64 - uint64(leftover))
		leftover -= 31
		staged >>= 33
		split32 = append(split32, uint32(staged))
	}

	if leftover > 0 {
		staged = curr << (64 - uint64(leftover))
		leftover -= 31
		staged |= next >> uint64(31+leftover)
		leftover += 64
		staged >>= 33

		split32 = append(split32, uint32(staged))
	}

	for leftover > 0 {
		staged = next << (64 - uint64(leftover))
		leftover -= 31
		staged >>= 33
		split32 = append(split32, uint32(staged))
	}

	return split32
}
