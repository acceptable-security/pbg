package disasm;

import (
	"github.com/bnagy/gapstone"
	"pbg/graph"
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
)

func getBinaryEntryPoint(pbg *graph.ProgramBehaviorGraph, binaryObj string) uint64 {
	entryPoints, err := pbg.Query(fmt.Sprintf("g.V('%s').Out('prog-entry-point').All()", binaryObj))

	if err != nil {
		panic(err)
	}

	if len(entryPoints) == 0 {
		panic(fmt.Sprintf("No entry points for binary %s", binaryObj))
	}

	entryPoint, err := strconv.ParseUint(entryPoints[0][1:len(entryPoints[0])-1], 16, 64)

	if err != nil {
		panic(err)
	}

	return entryPoint
}

func getBinarySectionAddr(pbg *graph.ProgramBehaviorGraph, binaryObj string, section string) uint64 {
	dataObjs, err := pbg.Query(fmt.Sprintf("g.V('%s').Out('has-section').Filter(regex('\\.%s')).Out('elf-section-addr').All()", binaryObj, section))

	if err != nil {
		panic(err)
	} 

	if len(dataObjs) != 1 {
		panic("Invalid data found")
	}

	data := dataObjs[0]
	data = data[1:len(data) - 1]

	addr, err := strconv.ParseUint(data, 16, 64)

	if err != nil {
		panic(err)
	}

	return addr
}

func getBinarySectionData(pbg *graph.ProgramBehaviorGraph, binaryObj string, section string) []byte {
	dataObjs, err := pbg.Query(fmt.Sprintf("g.V('%s').Out('has-section').Filter(regex('\\.%s')).Out('section-has-data').All()", binaryObj, section))

	if err != nil {
		panic(err)
	} 

	if len(dataObjs) != 1 {
		panic("Invalid data found")
	}

	data := dataObjs[0][1:len(dataObjs[0]) - 1]

	decoded_data, err := base64.StdEncoding.DecodeString(data)

	if err != nil {
		panic(err);
	}

	return decoded_data
}

func loadBinary(pbg *graph.ProgramBehaviorGraph, binaryObj string) {
	entryAddr := getBinaryEntryPoint(pbg, binaryObj)
	address := getBinarySectionAddr(pbg, binaryObj, "text")
	binaryData := getBinarySectionData(pbg, binaryObj, "text")

	log.Printf("Found text at 0x%x entry 0x%x from binary %s (%d bytes)\n", address, entryAddr, binaryObj, len(binaryData))

	engine, err := gapstone.New(
		gapstone.CS_ARCH_X86,
		gapstone.CS_MODE_64,
	)

	defer engine.Close()

	insns, err := engine.Disasm(binaryData, address, 0)

	if err != nil {
		panic(err);
	}

	pbg.AddRelationFunc(func (ch chan []string) {
		for _, insn := range insns {
			addrStr := fmt.Sprintf("0x%x", insn.Address)
			opStr := fmt.Sprintf("%s %s", insn.Mnemonic, insn.OpStr)
			ch <- []string { addrStr, "disassembles-to", opStr }
		}

		close(ch)
	})

	log.Printf("Handled %d instructions\n", len(insns))
}

func loadDisasm(pbg *graph.ProgramBehaviorGraph, opt map[string] interface{}) {
	binaryObjs, err := pbg.Query("g.V().In('prog-entry-point').All()");

	if err != nil {
		panic(err)
	}

	for _, binaryObj := range binaryObjs {
		loadBinary(pbg, binaryObj[1:len(binaryObj) - 1])
	}

	if err != nil {
		panic(err)
	}
}

func init() {
	graph.RegisterProvider("disasm", loadDisasm, "elf")
}