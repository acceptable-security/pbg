package elf

import (
	"fmt"
	"log"
	"os"
	"pbg/graph"
	"debug/elf"
	"encoding/base64"
	"strconv"
);

func loadElf(pbg *graph.ProgramBehaviorGraph, opt map[string] interface{}) {
	binaryObj, ok := opt["binary"].(string);

	// If none found, ignore
	if !ok {
		return;
	}

	file, err := os.Open(binaryObj)

	if err != nil {
		panic("Failed to open binary object for elf")
	}

	elfobj, err := elf.NewFile(file)

	if err != nil {
		panic("Invalid elf object")
	}

	log.Printf("Entry Point: from %s 0x%08x\n", binaryObj, elfobj.Entry)

	pbg.AddRelation(binaryObj, "prog-entry-point", fmt.Sprintf("%08x", elfobj.Entry));

	for _, section := range elfobj.Sections {
		sectionName := section.SectionHeader.Name
		sectionAddr := strconv.FormatUint(section.SectionHeader.Addr, 16)
		_sectionSize := section.SectionHeader.Size
		sectionSize := strconv.FormatUint(_sectionSize, 16)

		pbg.AddRelation(binaryObj, "has-section", sectionName);
		pbg.AddRelation(sectionName, "elf-section-size", sectionSize)
		pbg.AddRelation(sectionName, "elf-section-addr", sectionAddr)

		data := make([]byte, _sectionSize);
		n, err := section.ReaderAt.ReadAt(data, 0);

		if err != nil {
			log.Printf("Failed to get section %s\n", sectionName);
			continue;
		}

		if uint64(n) != _sectionSize {
			panic("Failed to read all bytes");
		}

		log.Printf("Reading section %s length %s at 0x%s\n", sectionName, sectionSize, sectionAddr);

		encoded_data := base64.StdEncoding.EncodeToString(data)
		pbg.AddRelation(sectionName, "section-has-data", string(encoded_data));
	}

	dwarfobj, err := elfobj.DWARF()

	if err != nil {
		log.Printf("Failed to find dwarf data (%v) skipping...", err)
		return
	}

	readDwarf(pbg, dwarfobj)
}

func init() {
	graph.RegisterProvider("elf", loadElf)
}