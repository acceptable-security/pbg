/* **********************************************************
 * Copyright (c) 2011-2018 Google, Inc.  All rights reserved.
 * **********************************************************/

/*
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * * Redistributions of source code must retain the above copyright notice,
 *   this list of conditions and the following disclaimer.
 *
 * * Redistributions in binary form must reproduce the above copyright notice,
 *   this list of conditions and the following disclaimer in the documentation
 *   and/or other materials provided with the distribution.
 *
 * * Neither the name of Google, Inc. nor the names of its contributors may be
 *   used to endorse or promote products derived from this software without
 *   specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL GOOGLE, INC. OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
 * SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
 * CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
 * LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
 * OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH
 * DAMAGE.
 */

#include <stdio.h>
#include <stddef.h>
#include "dr_api.h"
#include "drmgr.h"
#include "drwrap.h"
#include "drreg.h"

#define NULL_TERMINATE(buf) (buf)[(sizeof((buf)) / sizeof((buf)[0])) - 1] = '\0'
#define INIT_OK(EXPR) do { bool ok = (EXPR); DR_ASSERT(ok) } while(0)

// Each ins_ref_t describes an executed instruction.
typedef struct _ins_ref_t {
    app_pc pc;
} ins_ref_t;

/* Max number of ins_ref a buffer can have. It should be big enough
 * to hold all entries between clean calls.
 */
#define MAX_NUM_INS_REFS 8192

// The maximum size of buffer for holding ins_refs. 
#define MEM_BUF_SIZE (sizeof(ins_ref_t) * MAX_NUM_INS_REFS)

// Thread local storage unit
typedef struct {
    byte* seg_base;
    ins_ref_t* buf_base;
} per_thread_t;

static client_id_t client_id;

/* Allocated TLS slot offsets */
enum {
    INSTRACE_TLS_OFFS_BUF_PTR,
    INSTRACE_TLS_COUNT, /* total number of TLS slots allocated */
};

static reg_id_t tls_seg;
static uint tls_offs;
static int tls_idx;

#define TLS_SLOT(tls_base, enum_val) (void **)((byte *)(tls_base) + tls_offs + (enum_val))
#define BUF_PTR(tls_base) *(ins_ref_t **)TLS_SLOT(tls_base, INSTRACE_TLS_OFFS_BUF_PTR)
#define MINSERT instrlist_meta_preinsert

static void malloc_wrap_pre(void *wrapcxt, OUT void **user_data);
static void malloc_wrap_post(void *wrapcxt, OUT void **user_data);


static
void instrace(void *drcontext) {
    per_thread_t *data;
    , *buf_ptr;

    data = drmgr_get_tls_field(drcontext, tls_idx);
    ins_ref_t* buf_ptr = BUF_PTR(data->seg_base);

    /* We use libc's fprintf as it is buffered and much faster than dr_fprintf
     * for repeated printing that dominates performance, as the printing does here.
     */
    for ( ins_ref_t* ins_ref = (ins_ref_t*) data->buf_base; ins_ref < buf_ptr; ins_ref++ ) {
        fprintf(STDERR, "%p\n", (ptr_uint_t)ins_ref->pc, decode_opcode_name(ins_ref->opcode));
    }

    BUF_PTR(data->seg_base) = data->buf_base;
}

/* clean_call dumps the memory reference info to the log file */
static
void clean_call(void) {
    void* drcontext = dr_get_current_drcontext();
    instrace(drcontext);
}

static
void insert_load_buf_ptr(
    void *drcontext,
    instrlist_t *ilist, instr_t *where, reg_id_t reg_ptr
) {
    dr_insert_read_raw_tls(drcontext, ilist, where, tls_seg,
                           tls_offs + INSTRACE_TLS_OFFS_BUF_PTR, reg_ptr);
}

static
void insert_update_buf_ptr(
    void* drcontext,
    instrlist_t* ilist, instr_t* where, reg_id_t reg_ptr,
    int adjust
) {
    MINSERT(
        ilist, where,
        XINST_CREATE_add(drcontext, opnd_create_reg(reg_ptr), OPND_CREATE_INT16(adjust)));
    dr_insert_write_raw_tls(drcontext, ilist, where, tls_seg,
                            tls_offs + INSTRACE_TLS_OFFS_BUF_PTR, reg_ptr);
}

static
void insert_save_pc(
    void* drcontext,
    instrlist_t* ilist, instr_t* where,
    reg_id_t base, reg_id_t scratch, app_pc pc
) {
    instrlist_insert_mov_immed_ptrsz(drcontext, (ptr_int_t) pc, opnd_create_reg(scratch),
                                     ilist, where, NULL, NULL);
    MINSERT(ilist, where,
            XINST_CREATE_store(drcontext,
                               OPND_CREATE_MEMPTR(base, offsetof(ins_ref_t, pc)),
                               opnd_create_reg(scratch)));
}

/* insert inline code to add an instruction entry into the buffer */
static
void instrument_instr(void* drcontext, instrlist_t* ilist, instr_t* where) {
    // We need two scratch registers
    reg_id_t reg_ptr, reg_tmp;

    if ( drreg_reserve_register(drcontext, ilist, where, NULL, &reg_ptr) != DRREG_SUCCESS ||
         drreg_reserve_register(drcontext, ilist, where, NULL, &reg_tmp) != DRREG_SUCCESS ) {
        DR_ASSERT(false); // cannot recover
        return;
    }

    insert_load_buf_ptr(drcontext, ilist, where, reg_ptr);
    insert_save_pc(drcontext, ilist, where, reg_ptr, reg_tmp, instr_get_app_pc(where));
    insert_update_buf_ptr(drcontext, ilist, where, reg_ptr, sizeof(ins_ref_t));

    // Restore scratch registers 
    if (drreg_unreserve_register(drcontext, ilist, where, reg_ptr) != DRREG_SUCCESS ||
        drreg_unreserve_register(drcontext, ilist, where, reg_tmp) != DRREG_SUCCESS)
        DR_ASSERT(false);
}

// For each app instr, we insert inline code to fill the buffer.
static
dr_emit_flags_t event_app_instruction(
    void* drcontext, void* tag,
    instrlist_t* bb, instr_t* instr,
    bool for_trace, bool translating, void* user_data
) {
    // We don't want to auto-predicate any instrumentation 
    drmgr_disable_auto_predication(drcontext, bb);

    if ( !instr_is_app(instr) ) {
        return DR_EMIT_DEFAULT;
    }

    // Insert code to add an entry to the buffer
    instrument_instr(drcontext, bb, instr);

    // Insert code once per bb to call clean_call for processing the buffer
    if ( drmgr_is_first_instr(drcontext, instr) ) {
        dr_insert_clean_call(drcontext, bb, instr, (void*) clean_call, false, 0);
    }

    return DR_EMIT_DEFAULT;
}

static
void event_thread_init(void* drcontext) {
    per_thread_t* data = dr_thread_alloc(drcontext, sizeof(per_thread_t));
    DR_ASSERT(data != NULL);
    drmgr_set_tls_field(drcontext, tls_idx, data);

    /* Keep seg_base in a per-thread data structure so we can get the TLS
     * slot and find where the pointer points to in the buffer.
     */
    data->seg_base = dr_get_dr_segment_base(tls_seg);
    data->buf_base = dr_raw_mem_alloc(MEM_BUF_SIZE, DR_MEMPROT_READ | DR_MEMPROT_WRITE, NULL);

    DR_ASSERT(data->seg_base != NULL && data->buf_base != NULL);

    // Put buf_base to TLS as starting buf_ptr
    BUF_PTR(data->seg_base) = data->buf_base;
}

static
void module_load_event(void* drcontext, const module_data_t* mod, bool loaded) {
    app_pc towrap = (app_pc) dr_get_proc_address(mod->handle, "malloc");

    if ( towrap != NULL ) {
        bool ok = drwrap_wrap(towrap, malloc_wrap_pre, malloc_wrap_post);
        DR_ASSERT(ok);
    }
}

static
void module_exit_event(void) {
    INIT_OK(dr_raw_tls_cfree(tls_offs, INSTRACE_TLS_COUNT));
    INIT_OK(drmgr_unregister_tls_field(tls_idx));
    INIT_OK(drmgr_unregister_thread_init_event(event_thread_init));
    INIT_OK(drmgr_unregister_thread_exit_event(event_thread_exit));
    INIT_OK(drmgr_unregister_bb_insertion_event(event_app_instruction));

    drreg_status_t regok = drreg_exit(); DR_ASSERT(regok != DRREG_SUCCESS);

    drwrap_exit();
    drmgr_exit();
}

DR_EXPORT
void dr_client_main(client_id_t id, int argc, const char* argv[]) {
    dr_set_client_name("PBG Instructions & Heap Client", "https://github.com/block8437/pbg/issues");

    // make it easy to tell, by looking at log file, which client executed 
    dr_log(NULL, DR_LOG_ALL, 1, "Client instr_alloc initializing\n");

    // also give notification to stderr 
    if ( dr_is_notify_on() ) {
        dr_fprintf(STDERR, "Client instr_alloc is running\n");
    }

    // Initialize extensions
    INIT_OK(drmgr_init());
    INIT_OK(drwrap_init());

    // Register load/exit events
    INIT_OK(drmgr_register_module_load_event(module_load_event));
    INIT_OK(dr_register_exit_event(module_exit_event));
    INIT_OK(drmgr_register_thread_init_event(event_thread_init));
    INIT_OK(drmgr_register_thread_exit_event(event_thread_exit));
    INIT_OK(drmgr_register_bb_instrumentation_event(NULL, event_app_instruction, NULL));

    // Initialize drreg and reserve 2 registers
    drreg_options_t ops = { sizeof(ops), 3, false };
    drreg_status_t regok = drreg_init(&ops);
    DR_ASSERT(regok == DRREG_SUCCESS);

    // Get TLS index
    tls_idx = drmgr_register_tls_field();
    DR_ASSERT(tls_idx != -1);

    /* The TLS field provided by DR cannot be directly accessed from the code cache.
     * For better performance, we allocate raw TLS so that we can directly
     * access and update it with a single instruction.
     */
    INIT_OK(dr_raw_tls_calloc(&tls_seg, &tls_offs, INSTRACE_TLS_COUNT, 0));
}

static
void malloc_wrap_pre(void* wrapcxt, OUT void** user_data) {
    // malloc(size) or HeapAlloc(heap, flags, size) 
    size_t sz = (size_t) drwrap_get_arg(wrapcxt, 0);

    // find the maximum malloc request 
    dr_fprintf(STDERR, "malloc %zu", sz);

    *user_data = (void *)sz;
}

static
void malloc_wrap_post(void* wrapcxt, OUT void** user_data) {
    ptr_int_t ptr = (ptr_int_t) drwrap_get_retval(wrapcxt);
    dr_fprintf(STDERR, "%p\n", wrapcxt);
}
