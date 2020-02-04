package elf

import (
	"pbg/graph"
	"fmt"
	"io"
	"log"
	"encoding/binary"
	"debug/dwarf"
	"ekyu.moe/leb128"
)

func dwarfOffsetId(offset int64) string {
	return fmt.Sprintf("offset-%d", offset)
}

func dwarfTypeId(offset dwarf.Offset) string {
	return fmt.Sprintf("dwarf-type-%d", offset)
}

// Parse a DW_AT_location
func readDwarfLocation(pbg *graph.ProgramBehaviorGraph, varName string, data []byte, dwarfReader *dwarf.Reader) (string, error) {
	output := ""

	for len(data) > 0 {
		opcode := DwarfOpcode(data[0])

		if opcode == DW_OP_fbreg {
			// Framebuffer relative
			data = data[1:]
			offset, n := leb128.DecodeSleb128(data)
			data = data[n:]

			tmpStr := "+"

			if offset < 0 {
				tmpStr = "-";
				offset = -offset;
			}

			output += fmt.Sprintf("%s%s0x%x", DWARF_X86_SP.String(), tmpStr, offset)
		} else if opcode == DW_OP_addr {
			// Raw address
			data = data[1:]
			addrSize := dwarfReader.AddressSize()
			addr := binary.LittleEndian.Uint64(data[:addrSize])
			data = data[addrSize:]

			output += fmt.Sprintf("0x%08x", addr)
		} else {
			return "", fmt.Errorf("Unknown opcode %v (%v)", opcode.String(), data);
		}
	}

	return output, nil
}

// Parse a variable
func readDwarfVariable(pbg *graph.ProgramBehaviorGraph, entry *dwarf.Entry, dwarfReader *dwarf.Reader, prefix string) (string, error) {
	varNameField := entry.AttrField(dwarf.AttrName)

	// Ignore name-less variables.
	if varNameField == nil || varNameField.Val == nil {
		return "", fmt.Errorf("failed to find a name")
	}

	varName := varNameField.Val.(string)
	log.Printf("Found variable %s\n", varName)

	// Parse line number
	lineNumField := entry.AttrField(dwarf.AttrDeclLine)

	if lineNumField != nil && lineNumField.Val != nil {
		lineNum := lineNumField.Val.(int64)
		pbg.AddRelation(varName, "decl-at", fmt.Sprintf("line-%d", lineNum))
	}

	// Parse type id
	typeIdField := entry.AttrField(dwarf.AttrType)

	if typeIdField != nil && typeIdField.Val != nil {
		typeId  := typeIdField.Val.(dwarf.Offset)
		pbg.AddRelation(varName, "has-var-type", dwarfTypeId(typeId))
	}

	// Parse runtime location
	locationField := entry.AttrField(dwarf.AttrLocation)

	if locationField != nil {
		if locationField.Class != dwarf.ClassExprLoc {
			return "", fmt.Errorf("Unable to handle location class %d", locationField.Class)
		}

		loc := locationField.Val.([]byte)

		if locStr, err := readDwarfLocation(pbg, varName, loc, dwarfReader); err == nil {
			log.Printf("Found variable %s at %s\n", varName, locStr)
			pbg.AddRelation(varName, "runtime-at", locStr)
		} else {
			log.Printf("Failed to handle location: %v\n", err)
		}
	} else {
		log.Printf("Variable doesn't have location field\n")
	}

	return varName, nil
}

// Parse a parameter
func readDwarfParameter(pbg *graph.ProgramBehaviorGraph, entry *dwarf.Entry, dwarfReader *dwarf.Reader, prefix string) (string, error) {
	// TODO - not this
	return readDwarfVariable(pbg, entry, dwarfReader, prefix)
}

// Parse a lexical block/function block
func readDwarfBlock(pbg *graph.ProgramBehaviorGraph, entry *dwarf.Entry, dwarfReader *dwarf.Reader, funcName string) {
	log.Println("Start parsing block")

	if entry.Children {
		for {
			entry, err := dwarfReader.Next()

			if entry == nil {
				break
			}

			if err != nil {
				panic(err)
			}

			log.Printf("Found a %s in block\n", entry.Tag.GoString())

			if entry.Tag == 0 {
				log.Println("End of block")
				break
			}

			if entry.Tag == dwarf.TagVariable {
				if varName, err := readDwarfVariable(pbg, entry, dwarfReader, funcName); err == nil {
					pbg.AddRelation(funcName, "has-var", varName)
				} else {
					log.Printf("Failed to handle variable: %v\n", err)
				}
			} else if entry.Tag == dwarf.TagStructType {
				readDwarfStructType(pbg, entry, dwarfReader)
			} else if entry.Tag == dwarf.TagFormalParameter {
				if paramName, err := readDwarfParameter(pbg, entry, dwarfReader, funcName); err == nil {
					pbg.AddRelation(funcName, "has-var", paramName)
					pbg.AddRelation(funcName, "has-param", paramName)
				} else {
					log.Printf("Failed to handle variable: %v\n", err)
				}
			} else if entry.Tag == dwarf.TagLexDwarfBlock {
				readDwarfBlock(pbg, entry, dwarfReader, funcName)
			} else if entry.Tag == dwarf.TagLabel {
				continue
			} else if entry.Tag == dwarf.TagSubprogram {
				readDwarfFunction(pbg, entry, dwarfReader)
			} else if entry.Tag == dwarf.TagUnspecifiedParameters {
				continue
			} else if entry.Tag == dwarf.TagPointerType {
				readDwarfPointerType(pbg, entry, dwarfReader)
			} else if entry.Tag == dwarf.TagArrayType {
				readDwarfArrayType(pbg, entry, dwarfReader)
			} else if entry.Tag == dwarf.TagEnumerationType {
				readDwarfEnumeration(pbg, entry, dwarfReader)
			} else if entry.Tag == dwarf.TagSubroutineType {
				readDwarfSubroutineType(pbg, entry, dwarfReader)
			} else {
				panic(fmt.Errorf("Unknown tag %v in block", entry.Tag))
			}
		}
	}


	log.Println("Done parsing block")
}

// Parse a function
func readDwarfFunction(pbg *graph.ProgramBehaviorGraph, entry *dwarf.Entry, dwarfReader *dwarf.Reader) (string, bool) {
	funcNameField := entry.AttrField(dwarf.AttrName)

	if funcNameField == nil || funcNameField.Val == nil {
		return "", false
	}

	funcName, ok := funcNameField.Val.(string)

	if !ok {
		panic(fmt.Sprintf("Failed to parse function name %v", funcNameField))
	}

	log.Printf("Found function %s\n", funcName);

	if entry.Children {
		readDwarfBlock(pbg, entry, dwarfReader, funcName)
	}

	return funcName, true
}

// Parse a member type
func readDwarfMemberType(pbg *graph.ProgramBehaviorGraph, entry *dwarf.Entry, dwarfReader *dwarf.Reader) string {
	dTypeId := dwarfTypeId(dwarf.Offset(entry.Offset))
	memberNameField := entry.AttrField(dwarf.AttrName)

	if memberNameField != nil && memberNameField.Val != nil {
		if memberName, ok := memberNameField.Val.(string); ok {
			pbg.AddRelation(dTypeId, "has-member-name", memberName)
		}
	}

	memberTypeField := entry.AttrField(dwarf.AttrType)
	memberType := dwarfTypeId(memberTypeField.Val.(dwarf.Offset))
	pbg.AddRelation(dTypeId, "has-member-type", memberType)

	dataLocField := entry.AttrField(dwarf.AttrDataMemberLoc)

	if dataLocField != nil {
		if dataLocFieldRef, ok := dataLocField.Val.(int64); ok {
			pbg.AddRelation(dTypeId, "has-data-offset", dwarfOffsetId(dataLocFieldRef))
		} else if dataLocFieldBlock, ok := dataLocField.Val.([]byte); ok {
			log.Printf("Unable to handle block: %v\n", dataLocFieldBlock)
		} else {
			panic(fmt.Sprintf("Unable to handle type %T in DML", dataLocField.Val))
		}
	}

	return dTypeId
}

// Parse a struct
func readDwarfStructType(pbg *graph.ProgramBehaviorGraph, entry *dwarf.Entry, dwarfReader *dwarf.Reader) string {
	structName := readDwarfBaseType(pbg, entry, dwarfReader)

	if entry.Children {
		for {
			entry, err := dwarfReader.Next()

			if entry == nil || entry.Tag == 0 {
				break
			}

			if err != nil {
				panic(err)
			}

			if entry.Tag == dwarf.TagMember {
				memberName := readDwarfMemberType(pbg, entry, dwarfReader)
				pbg.AddRelation(structName, "has-member", memberName)
			}
		}
	}

	return structName
}

// Parse a base type
func readDwarfBaseType(pbg *graph.ProgramBehaviorGraph, entry *dwarf.Entry, dwarfReader *dwarf.Reader) string {
	typeNameField := entry.AttrField(dwarf.AttrName)

	dTypeId := dwarfTypeId(dwarf.Offset(entry.Offset))

	if typeNameField != nil {
		typeName := typeNameField.Val.(string)
		pbg.AddRelation(dTypeId, "has-type-name", typeName)
		log.Printf("Found type %s\n", typeName)
	}

	return dTypeId
}

// Parse a typedef
func readDwarfTypedef(pbg *graph.ProgramBehaviorGraph, entry *dwarf.Entry, dwarfReader *dwarf.Reader) string {
	typeNameField := entry.AttrField(dwarf.AttrName)
	dTypeId := dwarfTypeId(dwarf.Offset(entry.Offset))

	if typeNameField != nil {
		typeName := typeNameField.Val.(string)
		pbg.AddRelation(dTypeId, "has-type-name", typeName)
		log.Printf("Found type %s\n", typeName)
	}

	if typeField := entry.AttrField(dwarf.AttrType); typeField != nil {
		otherTypeId := dwarfTypeId(typeField.Val.(dwarf.Offset))
		pbg.AddRelation(dTypeId, "has-real-type", otherTypeId)
	}

	return dTypeId
}

// Parse a constant type
func readDwarfConstType(pbg *graph.ProgramBehaviorGraph, entry *dwarf.Entry, dwarfReader *dwarf.Reader) string {
	dTypeId := dwarfTypeId(dwarf.Offset(entry.Offset))

	if typeField := entry.AttrField(dwarf.AttrType); typeField != nil {
		otherTypeId := dwarfTypeId(typeField.Val.(dwarf.Offset))
		log.Printf("Found type id %s", otherTypeId)
		pbg.AddRelation(dTypeId, "const-type", otherTypeId)
	}

	return dTypeId
}

// Parse a restricted type
func readDwarfRestrictType(pbg *graph.ProgramBehaviorGraph, entry *dwarf.Entry, dwarfReader *dwarf.Reader) string {
	dTypeId := dwarfTypeId(dwarf.Offset(entry.Offset))

	if typeField := entry.AttrField(dwarf.AttrType); typeField != nil {
		otherTypeId := dwarfTypeId(typeField.Val.(dwarf.Offset))
		log.Printf("Found type id %s", otherTypeId)
		pbg.AddRelation(dTypeId, "restrict-type", otherTypeId)
	}

	return dTypeId
}

// Parse a pointer type
func readDwarfPointerType(pbg *graph.ProgramBehaviorGraph, entry *dwarf.Entry, dwarfReader *dwarf.Reader) string {
	dTypeId := dwarfTypeId(dwarf.Offset(entry.Offset))

	if typeField := entry.AttrField(dwarf.AttrType); typeField != nil {
		otherTypeId := dwarfTypeId(typeField.Val.(dwarf.Offset))
		log.Printf("Found type id %s", otherTypeId)
		pbg.AddRelation(dTypeId, "pointer-type", otherTypeId)
	}

	return dTypeId
}

// Parse an array type
func readDwarfArrayType(pbg *graph.ProgramBehaviorGraph, entry *dwarf.Entry, dwarfReader *dwarf.Reader) string {
	typeNameField := entry.AttrField(dwarf.AttrName)
	dTypeId := dwarfTypeId(dwarf.Offset(entry.Offset))

	if typeNameField != nil {
		typeName := typeNameField.Val.(string)
		pbg.AddRelation(dTypeId, "has-type-name", typeName)
		log.Printf("Found type %s\n", typeName)
	}

	if typeField := entry.AttrField(dwarf.AttrType); typeField != nil {
		otherTypeId := dwarfTypeId(typeField.Val.(dwarf.Offset))
		pbg.AddRelation(dTypeId, "array-type", otherTypeId)
	}

	if entry.Children {
		for {
			entry, err := dwarfReader.Next()

			if entry == nil {
				break
			}

			if err != nil {
				panic(err)
			}

			log.Printf("Found a %s in array\n", entry.Tag.GoString())

			if entry.Tag == dwarf.TagSubrangeType {
				subRange := readDwarfSubrangeType(pbg, entry, dwarfReader)
				pbg.AddRelation(dTypeId, "has-subrange", subRange)
			} else if entry.Tag == 0 {
				break
			} else {
				panic(fmt.Errorf("Unknown tag %v in array type", entry.Tag))
			}
		}
	}
	
	return dTypeId
}

// Parse a subrange type
func readDwarfSubrangeType(pbg *graph.ProgramBehaviorGraph, entry *dwarf.Entry, dwarfReader *dwarf.Reader) string {
	typeNameField := entry.AttrField(dwarf.AttrName)
	dTypeId := dwarfTypeId(dwarf.Offset(entry.Offset))

	if typeNameField != nil {
		typeName := typeNameField.Val.(string)
		pbg.AddRelation(dTypeId, "has-type-name", typeName)
		log.Printf("Found type %s\n", typeName)
	}

	if typeField := entry.AttrField(dwarf.AttrType); typeField != nil {
		otherTypeId := dwarfTypeId(typeField.Val.(dwarf.Offset))
		pbg.AddRelation(dTypeId, "enumerator-type", otherTypeId)
	}

	return dTypeId
}

// Parse an enumerator type
func readDwarfEnumeratorType(pbg *graph.ProgramBehaviorGraph, entry *dwarf.Entry, dwarfReader *dwarf.Reader) string {
	typeNameField := entry.AttrField(dwarf.AttrName)
	dTypeId := dwarfTypeId(dwarf.Offset(entry.Offset))

	if typeNameField != nil {
		typeName := typeNameField.Val.(string)
		pbg.AddRelation(dTypeId, "has-type-name", typeName)
		log.Printf("Found type %s\n", typeName)
	}

	if typeField := entry.AttrField(dwarf.AttrType); typeField != nil {
		otherTypeId := dwarfTypeId(typeField.Val.(dwarf.Offset))
		pbg.AddRelation(dTypeId, "subrange-type", otherTypeId)
	}

	return dTypeId
}

// Parse an enumeration
func readDwarfEnumeration(pbg *graph.ProgramBehaviorGraph, entry *dwarf.Entry, dwarfReader *dwarf.Reader) string {
	typeNameField := entry.AttrField(dwarf.AttrName)
	dTypeId := dwarfTypeId(dwarf.Offset(entry.Offset))

	if typeNameField != nil {
		typeName := typeNameField.Val.(string)
		pbg.AddRelation(dTypeId, "has-type-name", typeName)
		log.Printf("Found type %s\n", typeName)
	}

	if typeField := entry.AttrField(dwarf.AttrType); typeField != nil {
		otherTypeId := dwarfTypeId(typeField.Val.(dwarf.Offset))
		pbg.AddRelation(dTypeId, "enumeration-type", otherTypeId)
	}

	if entry.Children {
		for {
			entry, err := dwarfReader.Next()

			if entry == nil {
				break
			}

			if err != nil {
				panic(err)
			}

			log.Printf("Found a %s in enumeration\n", entry.Tag.GoString())

			if entry.Tag == dwarf.TagEnumerator {
				subRange := readDwarfEnumeratorType(pbg, entry, dwarfReader)
				pbg.AddRelation(dTypeId, "has-enumerator", subRange)
			} else if entry.Tag == 0 {
				break
			} else {
				panic(fmt.Errorf("Unknown tag %v in array type", entry.Tag))
			}
		}
	}

	return dTypeId
}

// Parse a subroutine type
func readDwarfSubroutineType(pbg *graph.ProgramBehaviorGraph, entry *dwarf.Entry, dwarfReader *dwarf.Reader) string {
	typeNameField := entry.AttrField(dwarf.AttrName)
	dTypeId := dwarfTypeId(dwarf.Offset(entry.Offset))

	if typeNameField != nil {
		typeName := typeNameField.Val.(string)
		pbg.AddRelation(dTypeId, "has-type-name", typeName)
		log.Printf("Found type %s\n", typeName)
	}

	if typeField := entry.AttrField(dwarf.AttrType); typeField != nil {
		otherTypeId := dwarfTypeId(typeField.Val.(dwarf.Offset))
		pbg.AddRelation(dTypeId, "subroutine-type", otherTypeId)
	}

	if entry.Children {
		for {
			entry, err := dwarfReader.Next()

			if entry == nil {
				break
			}

			if err != nil {
				panic(err)
			}

			log.Printf("Found a %s in subroutine\n", entry.Tag.GoString())

			if entry.Tag == dwarf.TagFormalParameter {
				if paramName, err := readDwarfParameter(pbg, entry, dwarfReader, "<>"); err == nil {
					pbg.AddRelation(dTypeId, "has-var", paramName)
					pbg.AddRelation(dTypeId, "has-param", paramName)
				} else {
					log.Printf("Failed to handle variable: %v\n", err)
				}
			} else if entry.Tag == 0 {
				break
			} else {
				panic(fmt.Errorf("Unknown tag %v in subroutine type", entry.Tag))
			}
		}
	}

	return dTypeId
}

// Parse a LineEntry
func readDwarfCULine(pbg *graph.ProgramBehaviorGraph, cuName string, lineReader *dwarf.LineReader) {
	var entry dwarf.LineEntry

	for {
		err := lineReader.Next(&entry)

		if err  != nil {
			if err == io.EOF {
				break
			}

			panic(err)
		}

		locName := cuName + ":" + fmt.Sprintf("line-%d", entry.Line)
		pbg.AddRelation(locName, "text-at-pc", fmt.Sprintf("0x%08x", entry.Address))
	}
}

// Parse a compile unit
func readDwarfCU(pbg *graph.ProgramBehaviorGraph, entry *dwarf.Entry, dwarfReader *dwarf.Reader) {
	cuNameField := entry.AttrField(dwarf.AttrName)
	cuName := cuNameField.Val.(string)

	if entry.Children {
		for {
			entry, err := dwarfReader.Next()

			if entry == nil {
				break
			}

			if err != nil {
				panic(err)
			}

			log.Printf("Found a %s in compilation unit\n", entry.Tag.GoString())

			if entry.Tag == dwarf.TagSubprogram {
				if funcName, ok := readDwarfFunction(pbg, entry, dwarfReader); ok {
					pbg.AddRelation(cuName, "defined-in", funcName);
				}
			} else if entry.Tag == dwarf.TagBaseType {
				readDwarfBaseType(pbg, entry, dwarfReader)
			} else if entry.Tag == dwarf.TagStructType || entry.Tag == dwarf.TagUnionType {
				readDwarfStructType(pbg, entry, dwarfReader)
			} else if entry.Tag == dwarf.TagTypedef {
				readDwarfTypedef(pbg, entry, dwarfReader)
			} else if entry.Tag == dwarf.TagConstType {
				readDwarfConstType(pbg, entry, dwarfReader)
			} else if entry.Tag == dwarf.TagRestrictType {
				readDwarfRestrictType(pbg, entry, dwarfReader)
			} else if entry.Tag == dwarf.TagPointerType {
				readDwarfPointerType(pbg, entry, dwarfReader)
			} else if entry.Tag == dwarf.TagArrayType {
				readDwarfArrayType(pbg, entry, dwarfReader)
			} else if entry.Tag == dwarf.TagEnumerationType {
				readDwarfEnumeration(pbg, entry, dwarfReader)
			} else if entry.Tag == dwarf.TagSubroutineType {
				readDwarfSubroutineType(pbg, entry, dwarfReader)
			} else if entry.Tag == dwarf.TagVariable {
				if varName, err := readDwarfVariable(pbg, entry, dwarfReader, "<global>"); err == nil {
					pbg.AddRelation(cuName, "has-global-var", varName)
				} else {
					log.Printf("Failed to handle variable: %v\n", err)
				}
			} else if entry.Tag == dwarf.TagFormalParameter {
				// This happens to a subprogram thats declared at the end of a function
				// which has no sibling
				log.Printf("Dangling parameter, but can't handle")
				continue
			} else if entry.Tag == 0 {
				break
			} else {
				panic(fmt.Errorf("Unknown tag %v in CU", entry.Tag))
			}
		}
	}
}

func readDwarf(pbg *graph.ProgramBehaviorGraph, dwarfObj *dwarf.Data) {
	dwarfReader := dwarfObj.Reader()

	for {
		entry, err := dwarfReader.Next()

		if entry == nil {
			break
		}

		if err != nil {
			panic(err)
		}

		log.Printf("Found a %s\n", entry.Tag.GoString())

		if entry.Tag == dwarf.TagCompileUnit {
			lineReader, err := dwarfObj.LineReader(entry)

			if err != nil {
				panic(err)
			}

		readDwarfCULine(pbg, entry.AttrField(dwarf.AttrName).Val.(string), lineReader)
			readDwarfCU(pbg, entry, dwarfReader)
		} else if entry.Tag == dwarf.TagBaseType {
			panic(fmt.Errorf("Unknown tag %v at root", entry.Tag))
		}
	}
}
