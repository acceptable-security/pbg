package elf

// "Borrowed" from pyelftools
type DwarfOpcode uint
const (
	DW_OP_addr DwarfOpcode                = 0x03
	DW_OP_deref DwarfOpcode               = 0x06
	DW_OP_const1u DwarfOpcode             = 0x08
	DW_OP_const1s DwarfOpcode             = 0x09
	DW_OP_const2u DwarfOpcode             = 0x0a
	DW_OP_const2s DwarfOpcode             = 0x0b
	DW_OP_const4u DwarfOpcode             = 0x0c
	DW_OP_const4s DwarfOpcode             = 0x0d
	DW_OP_const8u DwarfOpcode             = 0x0e
	DW_OP_const8s DwarfOpcode             = 0x0f
	DW_OP_constu DwarfOpcode              = 0x10
	DW_OP_consts DwarfOpcode              = 0x11
	DW_OP_dup DwarfOpcode                 = 0x12
	DW_OP_drop DwarfOpcode                = 0x13
	DW_OP_over DwarfOpcode                = 0x14
	DW_OP_pick DwarfOpcode                = 0x15
	DW_OP_swap DwarfOpcode                = 0x16
	DW_OP_rot DwarfOpcode                 = 0x17
	DW_OP_xderef DwarfOpcode              = 0x18
	DW_OP_abs DwarfOpcode                 = 0x19
	DW_OP_and DwarfOpcode                 = 0x1a
	DW_OP_div DwarfOpcode                 = 0x1b
	DW_OP_minus DwarfOpcode               = 0x1c
	DW_OP_mod DwarfOpcode                 = 0x1d
	DW_OP_mul DwarfOpcode                 = 0x1e
	DW_OP_neg DwarfOpcode                 = 0x1f
	DW_OP_not DwarfOpcode                 = 0x20
	DW_OP_or DwarfOpcode                  = 0x21
	DW_OP_plus DwarfOpcode                = 0x22
	DW_OP_plus_uconst DwarfOpcode         = 0x23
	DW_OP_shl DwarfOpcode                 = 0x24
	DW_OP_shr DwarfOpcode                 = 0x25
	DW_OP_shra DwarfOpcode                = 0x26
	DW_OP_xor DwarfOpcode                 = 0x27
	DW_OP_bra DwarfOpcode                 = 0x28
	DW_OP_eq DwarfOpcode                  = 0x29
	DW_OP_ge DwarfOpcode                  = 0x2a
	DW_OP_gt DwarfOpcode                  = 0x2b
	DW_OP_le DwarfOpcode                  = 0x2c
	DW_OP_lt DwarfOpcode                  = 0x2d
	DW_OP_ne DwarfOpcode                  = 0x2e
	DW_OP_skip DwarfOpcode                = 0x2f
	DW_OP_regx DwarfOpcode                = 0x90
	DW_OP_fbreg DwarfOpcode               = 0x91
	DW_OP_bregx DwarfOpcode               = 0x92
	DW_OP_piece DwarfOpcode               = 0x93
	DW_OP_deref_size DwarfOpcode          = 0x94
	DW_OP_xderef_size DwarfOpcode         = 0x95
	DW_OP_nop DwarfOpcode                 = 0x96
	DW_OP_push_object_address DwarfOpcode = 0x97
	DW_OP_call2 DwarfOpcode               = 0x98
	DW_OP_call4 DwarfOpcode               = 0x99
	DW_OP_call_ref DwarfOpcode            = 0x9a
	DW_OP_form_tls_address DwarfOpcode    = 0x9b
	DW_OP_call_frame_cfa DwarfOpcode      = 0x9c
	DW_OP_bit_piece DwarfOpcode           = 0x9d
	DW_OP_implicit_value DwarfOpcode      = 0x9e
	DW_OP_stack_value DwarfOpcode         = 0x9f
	DW_OP_implicit_pointer DwarfOpcode    = 0xa0
	DW_OP_addrx DwarfOpcode               = 0xa1
	DW_OP_constx DwarfOpcode              = 0xa2
	DW_OP_entry_value DwarfOpcode         = 0xa3
	DW_OP_const_type DwarfOpcode          = 0xa4
	DW_OP_regval_type DwarfOpcode         = 0xa5
	DW_OP_deref_type DwarfOpcode          = 0xa6
	DW_OP_xderef_type DwarfOpcode         = 0xa7
	DW_OP_convert DwarfOpcode             = 0xa8
	DW_OP_reinterpret DwarfOpcode         = 0xa9
	DW_OP_lo_user DwarfOpcode             = 0xe0
	DW_OP_hi_user DwarfOpcode             = 0xff
)

func (reg DwarfOpcode) String() string {
	if ( reg == DW_OP_addr ) { return "addr"; }
	if ( reg == DW_OP_deref ) { return "deref"; }
	if ( reg == DW_OP_const1u ) { return "const1u"; }
	if ( reg == DW_OP_const1s ) { return "const1s"; }
	if ( reg == DW_OP_const2u ) { return "const2u"; }
	if ( reg == DW_OP_const2s ) { return "const2s"; }
	if ( reg == DW_OP_const4u ) { return "const4u"; }
	if ( reg == DW_OP_const4s ) { return "const4s"; }
	if ( reg == DW_OP_const8u ) { return "const8u"; }
	if ( reg == DW_OP_const8s ) { return "const8s"; }
	if ( reg == DW_OP_constu ) { return "constu"; }
	if ( reg == DW_OP_consts ) { return "consts"; }
	if ( reg == DW_OP_dup ) { return "dup"; }
	if ( reg == DW_OP_drop ) { return "drop"; }
	if ( reg == DW_OP_over ) { return "over"; }
	if ( reg == DW_OP_pick ) { return "pick"; }
	if ( reg == DW_OP_swap ) { return "swap"; }
	if ( reg == DW_OP_rot ) { return "rot"; }
	if ( reg == DW_OP_xderef ) { return "xderef"; }
	if ( reg == DW_OP_abs ) { return "abs"; }
	if ( reg == DW_OP_and ) { return "and"; }
	if ( reg == DW_OP_div ) { return "div"; }
	if ( reg == DW_OP_minus ) { return "minus"; }
	if ( reg == DW_OP_mod ) { return "mod"; }
	if ( reg == DW_OP_mul ) { return "mul"; }
	if ( reg == DW_OP_neg ) { return "neg"; }
	if ( reg == DW_OP_not ) { return "not"; }
	if ( reg == DW_OP_or ) { return "or"; }
	if ( reg == DW_OP_plus ) { return "plus"; }
	if ( reg == DW_OP_plus_uconst ) { return "plus_uconst"; }
	if ( reg == DW_OP_shl ) { return "shl"; }
	if ( reg == DW_OP_shr ) { return "shr"; }
	if ( reg == DW_OP_shra ) { return "shra"; }
	if ( reg == DW_OP_xor ) { return "xor"; }
	if ( reg == DW_OP_bra ) { return "bra"; }
	if ( reg == DW_OP_eq ) { return "eq"; }
	if ( reg == DW_OP_ge ) { return "ge"; }
	if ( reg == DW_OP_gt ) { return "gt"; }
	if ( reg == DW_OP_le ) { return "le"; }
	if ( reg == DW_OP_lt ) { return "lt"; }
	if ( reg == DW_OP_ne ) { return "ne"; }
	if ( reg == DW_OP_skip ) { return "skip"; }
	if ( reg == DW_OP_regx ) { return "regx"; }
	if ( reg == DW_OP_fbreg ) { return "fbreg"; }
	if ( reg == DW_OP_bregx ) { return "bregx"; }
	if ( reg == DW_OP_piece ) { return "piece"; }
	if ( reg == DW_OP_deref_size ) { return "deref_size"; }
	if ( reg == DW_OP_xderef_size ) { return "xderef_size"; }
	if ( reg == DW_OP_nop ) { return "nop"; }
	if ( reg == DW_OP_push_object_address ) { return "push_object_address"; }
	if ( reg == DW_OP_call2 ) { return "call2"; }
	if ( reg == DW_OP_call4 ) { return "call4"; }
	if ( reg == DW_OP_call_ref ) { return "call_ref"; }
	if ( reg == DW_OP_form_tls_address ) { return "form_tls_address"; }
	if ( reg == DW_OP_call_frame_cfa ) { return "call_frame_cfa"; }
	if ( reg == DW_OP_bit_piece ) { return "bit_piece"; }
	if ( reg == DW_OP_implicit_value ) { return "implicit_value"; }
	if ( reg == DW_OP_stack_value ) { return "stack_value"; }
	if ( reg == DW_OP_implicit_pointer ) { return "implicit_pointer"; }
	if ( reg == DW_OP_addrx ) { return "addrx"; }
	if ( reg == DW_OP_constx ) { return "constx"; }
	if ( reg == DW_OP_entry_value ) { return "entry_value"; }
	if ( reg == DW_OP_const_type ) { return "const_type"; }
	if ( reg == DW_OP_regval_type ) { return "regval_type"; }
	if ( reg == DW_OP_deref_type ) { return "deref_type"; }
	if ( reg == DW_OP_xderef_type ) { return "xderef_type"; }
	if ( reg == DW_OP_convert ) { return "convert"; }
	if ( reg == DW_OP_reinterpret ) { return "reinterpret"; }
	if ( reg == DW_OP_lo_user ) { return "lo_user"; }
	if ( reg == DW_OP_hi_user ) { return "hi_user"; }

	return "unknown"
}

type Dwarfx86Reg uint
type Dwarfx8664Reg uint

// Borrowed from the linux kernel
const (
	DWARF_X86_EAX Dwarfx86Reg = iota
	DWARF_X86_ECX
	DWARF_X86_EDX
	DWARF_X86_EBX
	DWARF_X86_ESP
	DWARF_X86_EBP
	DWARF_X86_ESI
	DWARF_X86_EDI
	DWARF_X86_EIP
	DWARF_X86_EFLAGS
	DWARF_X86_TRAPNO
	DWARF_X86_ST0
	DWARF_X86_ST1
	DWARF_X86_ST2
	DWARF_X86_ST3
	DWARF_X86_ST4
	DWARF_X86_ST5
	DWARF_X86_ST6
	DWARF_X86_ST
	DWARF_X86_CFA_REG_COLUMN
	DWARF_X86_CFA_OFF_COLUMN
	DWARF_X86_REGS_NUM
	DWARF_X86_SP = DWARF_X86_ESP

	// Reset iota for x86-64
	DWARF_X86_64_RAX Dwarfx8664Reg = iota
	DWARF_X86_64_RDX
	DWARF_X86_64_RCX
	DWARF_X86_64_RBX
	DWARF_X86_64_RSI
	DWARF_X86_64_RDI
	DWARF_X86_64_RBP
	DWARF_X86_64_RSP
	DWARF_X86_64_R8
	DWARF_X86_64_R9
	DWARF_X86_64_R10
	DWARF_X86_64_R11
	DWARF_X86_64_R12
	DWARF_X86_64_R13
	DWARF_X86_64_R14
	DWARF_X86_64_R15
	DWARF_X86_64_RIP
	DWARF_X86_64_CFA_REG_COLUMN
	DWARF_X86_64_CFA_OFF_COLUMN
	DWARF_X86_64_REGS_NUM
	DWARF_X86_64_SP = DWARF_X86_64_RSP
)

func (reg Dwarfx86Reg) String() string {
	names := []string{
		"EAX",
		"ECX",
		"EDX",
		"EBX",
		"ESP",
		"EBP",
		"ESI",
		"EDI",
		"EIP",
		"EFLAGS",
		"TRAPNO",
		"ST0",
		"ST1",
		"ST2",
		"ST3",
		"ST4",
		"ST5",
		"ST6",
		"ST",
		"CFA_REG_COLUMN",
		"CFA_OFF_COLUMN"}

	if reg < DWARF_X86_EAX || reg >= DWARF_X86_REGS_NUM {
		return "Unknown"
	}

	return names[reg]
}