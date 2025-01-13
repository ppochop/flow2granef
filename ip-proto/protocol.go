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
	num  uint16
	name string
}

type mapNum map[uint16]string
type mapName map[string]uint16

var rangeRE *regexp.Regexp

//go:embed protocol-numbers-1.csv
var protoNumbers string
var protoMappingNum mapNum
var protoMappingName mapName

//go:embed dns-parameters-4.csv
var rrtypeNumbers string
var rrtypeMappingNum mapNum
var rrtypeMappingName mapName

func init() {
	rangeRE, _ = regexp.Compile(`^(\d)+-(\d)+$`)
	protoMappingNum = mapNum{}
	protoMappingName = mapName{}
	rrtypeMappingNum = mapNum{}
	rrtypeMappingName = mapName{}
	rIp := csv.NewReader(strings.NewReader(protoNumbers))
	rRrtype := csv.NewReader(strings.NewReader(rrtypeNumbers))
	ipRecords, err := rIp.ReadAll()
	if err != nil {
		slog.Error("could not read IP protocol mappings", "error", err)
		panic(1)
	}
	rrtypeRecords, err := rRrtype.ReadAll()
	if err != nil {
		slog.Error("could not read DNS RRTYPE mappings", "error", err)
		panic(1)
	}
	ParseProtocolCSV(ipRecords[1:], 0, 1, protoMappingNum, protoMappingName)
	ParseProtocolCSV(rrtypeRecords[1:], 1, 0, rrtypeMappingNum, rrtypeMappingName)
}

func ParseProtocolCSV(records [][]string, numCol int, nameCol int, mNum mapNum, mName mapName) {
	for _, rec := range records {
		num, err := strconv.ParseUint(rec[numCol], 10, 16)
		if err != nil {
			tryRange := rangeRE.FindStringSubmatch(rec[numCol]) // maybe it's a range?
			if tryRange == nil {                                // not a range
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
				name := rec[nameCol]
				if name == "" {
					name = fmt.Sprintf("UNKNOWN(%d)", num)
				}
				nameUpper := strings.ToUpper(name)
				mNum[uint16(num)] = nameUpper
				mName[nameUpper] = uint16(fromNum)
			}
			// registered the whole range, move on
			continue
		}
		name := rec[nameCol]
		if name == "" {
			name = fmt.Sprintf("UNKNOWN(%d)", num)
		}
		nameUpper := strings.ToUpper(name)
		mNum[uint16(num)] = nameUpper
		mName[name] = uint16(num)
	}
}

func RRTypeFromName(name string) Protocol {
	nameUpper := strings.ToUpper(name)
	v, found := rrtypeMappingName[nameUpper]
	if !found {
		// TODO log
		return Protocol{
			name: nameUpper,
			num:  65535,
		}
	}

	return Protocol{
		name: nameUpper,
		num:  v,
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

func RRTypeFromNum(num uint16) Protocol {
	v, found := rrtypeMappingNum[num]
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

func ProtocolFromNum(num uint16) Protocol {
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

func (p *Protocol) GetNum() uint16 {
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
