package ipproto

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
)

type Protocol struct {
	num  uint8
	name string
}

//go:embed protocol-numbers-1.csv
var protoNumbers string
var protoMappingNum map[uint8]string
var protoMappingName map[string]uint8

func init() {
	re, _ := regexp.Compile(`^(\d)+-(\d)+$`)
	protoMappingNum = map[uint8]string{}
	protoMappingName = map[string]uint8{}
	r := csv.NewReader(strings.NewReader(protoNumbers))
	records, err := r.ReadAll()
	if err != nil {
		slog.Error("could not read protocol mappings", "error", err)
		panic(1)
	}
	for _, rec := range records[1:] {
		num, err := strconv.ParseUint(rec[0], 10, 8)
		if err != nil {
			tryRange := re.FindStringSubmatch(rec[0]) // maybe it's a range?
			if tryRange == nil {                      // not a range
				slog.Error("could not parse protocol number", "error", err)
				panic(1)
			}
			// get the bounds of the range
			fromNum, err := strconv.ParseUint(tryRange[1], 10, 8)
			if err != nil {
				slog.Error("could not parse protocol number from range", "error", err)
				panic(1)
			}
			toNum, err := strconv.ParseUint(tryRange[1], 10, 8)
			if err != nil {
				slog.Error("could not parse protocol number from range", "error", err)
				panic(1)
			}

			// register the range
			for i := fromNum; i <= toNum; i++ {
				name := rec[1]
				if name == "" {
					name = fmt.Sprintf("UNKNOWN(%d)", num)
				}
				nameUpper := strings.ToUpper(name)
				protoMappingNum[uint8(num)] = nameUpper
				protoMappingName[nameUpper] = uint8(fromNum)
			}
			// registered the whole range, move on
			continue
		}
		name := rec[1]
		if name == "" {
			name = fmt.Sprintf("UNKNOWN(%d)", num)
		}
		nameUpper := strings.ToUpper(name)
		protoMappingNum[uint8(num)] = nameUpper
		protoMappingName[name] = uint8(num)
	}
}

func ProtocolFromName(name string) Protocol {
	nameUpper := strings.ToUpper(name)
	v, found := protoMappingName[nameUpper]
	if !found {
		// TODO log
		return Protocol{
			name: nameUpper,
			num:  254,
		}
	}

	return Protocol{
		name: nameUpper,
		num:  v,
	}
}

func ProtocolFromNum(num uint8) Protocol {
	v, found := protoMappingNum[num]
	if !found {
		return Protocol{
			name: fmt.Sprintf("UNKNOWN(%d)", num),
			num:  num,
		}
	}

	return Protocol{
		name: v,
		num:  num,
	}
}

func (p *Protocol) GetNum() uint8 {
	return p.num
}

func (p *Protocol) GetName() string {
	return p.name
}

func (p *Protocol) IsIcmp() bool {
	return p.num == 1 || p.num == 58
}

func (p *Protocol) MarshalText() ([]byte, error) {
	return []byte(p.name), nil
}
