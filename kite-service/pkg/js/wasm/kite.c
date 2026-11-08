// Glue between QuickJS-NG and pkg/js. Built without quickjs-libc, so scripts
// get the language and what is defined here: no files, network or timers.
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#include "quickjs.h"

#define IMPORT(n) __attribute__((import_module("kite"), import_name(n)))
#define EXPORT(n) __attribute__((export_name(n)))

// Returns 1 once the run is out of time.
IMPORT("interrupt") int32_t host_interrupt(void);
// level: 0 log, 1 warn, 2 error.
IMPORT("log") void host_log(int32_t level, const char *p, uint32_t n);
// Runs a host function, returns the length of its JSON result, negative for
// an error message. host_take copies the pending result into dst.
IMPORT("call") int32_t host_call(const char *name, uint32_t name_len, const char *arg, uint32_t arg_len);
IMPORT("take") void host_take(char *dst);
// Starts a host call that Go answers later with kite_resolve. Returns its ID
// from 1, or the negative length of an error message like host_call.
IMPORT("start") int32_t host_start(const char *name, uint32_t name_len, const char *arg, uint32_t arg_len);
// QuickJS seeds Math.random from the clock when a context is created, so
// every instance restored from one snapshot would share the seed.
IMPORT("random") double host_random(void);

static JSRuntime *rt;
static JSContext *ctx;
static char *out;
static uint32_t out_len;

static int interrupt_handler(JSContext *c, void *opaque) {
  return host_interrupt();
}

// prelude.js passes one string with the arguments already joined.
static JSValue js_log(JSContext *ctx, JSValueConst this_val, int argc, JSValueConst *argv, int magic) {
  size_t len;
  const char *str = JS_ToCStringLen(ctx, &len, argv[0]);
  if (!str) return JS_EXCEPTION;
  host_log(magic, str, len);
  JS_FreeCString(ctx, str);
  return JS_UNDEFINED;
}

typedef int32_t (*host_fn)(const char *name, uint32_t name_len, const char *arg, uint32_t arg_len);

// Calls fn with the name in argv[0] and argv[1] as JSON, and stores what it
// returns in n. Returns -1 for an exception.
static int with_host_args(JSContext *ctx, int argc, JSValueConst *argv, host_fn fn, int32_t *n) {
  size_t name_len, arg_len = 4;
  const char *name = JS_ToCStringLen(ctx, &name_len, argv[0]);
  if (!name) return -1;
  JSValue arg = JS_JSONStringify(ctx, argc > 1 ? argv[1] : JS_NULL, JS_UNDEFINED, JS_UNDEFINED);
  if (JS_IsException(arg)) {
    JS_FreeCString(ctx, name);
    return -1;
  }
  const char *arg_str = "null";
  if (!JS_IsUndefined(arg)) {
    arg_str = JS_ToCStringLen(ctx, &arg_len, arg);
    if (!arg_str) {
      JS_FreeCString(ctx, name);
      JS_FreeValue(ctx, arg);
      return -1;
    }
  }

  *n = fn(name, name_len, arg_str, arg_len);
  JS_FreeCString(ctx, name);
  if (!JS_IsUndefined(arg)) JS_FreeCString(ctx, arg_str);
  JS_FreeValue(ctx, arg);
  return 0;
}

// Takes the result Go has pending: n bytes of JSON, or for a negative n an
// error message, which it throws as a TypeError for helpers like formatDate
// and as an Error for calls.
static JSValue take(JSContext *ctx, int32_t n, bool type_error) {
  uint32_t len = n < 0 ? -n : n;
  char *buf = js_malloc(ctx, len + 1);
  if (!buf) return JS_EXCEPTION;
  host_take(buf);
  buf[len] = 0;

  JSValue res;
  if (n >= 0) {
    res = JS_ParseJSON(ctx, buf, len, "<host>");
  } else if (type_error) {
    res = JS_ThrowTypeError(ctx, "%s", buf);
  } else {
    res = JS_ThrowPlainError(ctx, "%s", buf);
  }
  js_free(ctx, buf);
  return res;
}

// Helpers like formatDate run right away.
static JSValue js_call(JSContext *ctx, JSValueConst this_val, int argc, JSValueConst *argv) {
  int32_t n;
  if (with_host_args(ctx, argc, argv, host_call, &n) < 0) return JS_EXCEPTION;
  return take(ctx, n, true);
}

// The resolving functions of the calls Go hasn't answered yet, by ID. Go
// starts at most MAX_CALLS, from 1. Matches MaxCalls in js.go.
#define MAX_CALLS 20
static JSValue calls[MAX_CALLS + 1][2];

// Starts a call, which Go answers with kite_resolve, and returns its promise.
static JSValue js_start(JSContext *ctx, JSValueConst this_val, int argc, JSValueConst *argv) {
  // Before Go starts the call, so it can always be answered.
  JSValue funcs[2];
  JSValue promise = JS_NewPromiseCapability(ctx, funcs);
  if (JS_IsException(promise)) return promise;

  int32_t id;
  if (with_host_args(ctx, argc, argv, host_start, &id) < 0) {
    JS_FreeValue(ctx, promise);
    JS_FreeValue(ctx, funcs[0]);
    JS_FreeValue(ctx, funcs[1]);
    return JS_EXCEPTION;
  }

  if (id > 0 && id <= MAX_CALLS) {
    calls[id][0] = funcs[0];
    calls[id][1] = funcs[1];
    return promise;
  }

  // Go turned it down, like an unknown call.
  JS_FreeValue(ctx, take(ctx, id, false));
  JSValue err = JS_GetException(ctx);
  JSValue r = JS_Call(ctx, funcs[1], JS_UNDEFINED, 1, &err);
  JS_FreeValue(ctx, r);
  JS_FreeValue(ctx, err);
  JS_FreeValue(ctx, funcs[0]);
  JS_FreeValue(ctx, funcs[1]);
  return promise;
}

static JSValue js_random(JSContext *ctx, JSValueConst this_val, int argc, JSValueConst *argv) {
  return JS_NewFloat64(ctx, host_random());
}

static const JSCFunctionListEntry console_funcs[] = {
    JS_CFUNC_MAGIC_DEF("log", 1, js_log, 0),
    JS_CFUNC_MAGIC_DEF("info", 1, js_log, 0),
    JS_CFUNC_MAGIC_DEF("warn", 1, js_log, 1),
    JS_CFUNC_MAGIC_DEF("error", 1, js_log, 2),
};

// The output Go reads after kite_run, the result or the error.
static void set_out(const char *s, size_t len) {
  free(out);
  out = malloc(len);
  out_len = out ? len : 0;
  if (out) memcpy(out, s, len);
}

static void append_out(const char *s, size_t len) {
  char *grown = realloc(out, out_len + len);
  if (!grown) return;
  memcpy(grown + out_len, s, len);
  out = grown;
  out_len += len;
}

#define SET_OUT_LIT(s) set_out(s, sizeof(s) - 1)

// Stores the pending exception as the output, its message, a null byte and
// its stack.
static int32_t fail(void) {
  JSValue exc = JS_GetException(ctx);
  JSValue msg = JS_ToString(ctx, exc);
  size_t len;
  const char *m = NULL;
  if (JS_IsException(msg)) {
    // Like an object whose toString throws, or runs out of time.
    JSValue err = JS_GetException(ctx);
    if (JS_IsUncatchableError(err)) {
      SET_OUT_LIT("InternalError: interrupted");
    } else {
      SET_OUT_LIT("Error: the code threw something that can't be shown as text");
    }
    JS_FreeValue(ctx, err);
  } else if ((m = JS_ToCStringLen(ctx, &len, msg))) {
    set_out(m, len);
    JS_FreeCString(ctx, m);
  } else {
    SET_OUT_LIT("InternalError: out of memory");
  }
  JSValue stack = JS_IsError(exc) ? JS_GetPropertyStr(ctx, exc, "stack") : JS_UNDEFINED;
  const char *st = JS_IsString(stack) ? JS_ToCStringLen(ctx, &len, stack) : NULL;
  if (st) {
    append_out("", 1);
    append_out(st, len);
    JS_FreeCString(ctx, st);
  }
  JS_FreeValue(ctx, stack);
  JS_FreeValue(ctx, msg);
  JS_FreeValue(ctx, exc);
  return 1;
}

// Sets up QuickJS and runs the prelude, which sets up the globals scripts
// get and can remove kite.
EXPORT("kite_init")
int32_t kite_init(uint32_t memory_limit, uint32_t stack_size, const char *prelude, uint32_t prelude_len) {
  rt = JS_NewRuntime();
  if (!rt) return -1;
  JS_SetMemoryLimit(rt, memory_limit);
  JS_SetMaxStackSize(rt, stack_size);
  JS_SetInterruptHandler(rt, interrupt_handler, NULL);
  ctx = JS_NewContext(rt);
  if (!ctx) return -1;
  for (int i = 0; i <= MAX_CALLS; i++) {
    calls[i][0] = JS_UNDEFINED;
    calls[i][1] = JS_UNDEFINED;
  }

  JSValue global = JS_GetGlobalObject(ctx);
  JSValue kite = JS_NewObject(ctx);
  JS_SetPropertyStr(ctx, kite, "call", JS_NewCFunction(ctx, js_call, "call", 2));
  JS_SetPropertyStr(ctx, kite, "start", JS_NewCFunction(ctx, js_start, "start", 2));
  JS_SetPropertyStr(ctx, global, "kite", kite);
  JSValue console = JS_NewObject(ctx);
  JS_SetPropertyFunctionList(ctx, console, console_funcs, sizeof(console_funcs) / sizeof(console_funcs[0]));
  JS_SetPropertyStr(ctx, global, "console", console);
  JSValue math = JS_GetPropertyStr(ctx, global, "Math");
  JS_SetPropertyStr(ctx, math, "random", JS_NewCFunction(ctx, js_random, "random", 0));
  JS_FreeValue(ctx, math);
  JS_FreeValue(ctx, global);

  JSValue res = JS_Eval(ctx, prelude, prelude_len, "prelude.js", JS_EVAL_TYPE_GLOBAL | JS_EVAL_FLAG_STRICT);
  if (JS_IsException(res)) {
    fail();
    return -2;
  }
  JS_FreeValue(ctx, res);
  return 0;
}

EXPORT("kite_alloc")
void *kite_alloc(uint32_t n) { return malloc(n); }

EXPORT("kite_free")
void kite_free(void *p) { free(p); }

EXPORT("kite_out_ptr")
char *kite_out_ptr(void) { return out; }

EXPORT("kite_out_len")
uint32_t kite_out_len(void) { return out_len; }

// Logs an error nothing waits for at the error level.
static void log_error(JSValueConst err) {
  JSValue msg = JS_ToString(ctx, err);
  size_t len;
  const char *m = JS_IsException(msg) ? NULL : JS_ToCStringLen(ctx, &len, msg);
  if (m) {
    host_log(2, m, len);
    JS_FreeCString(ctx, m);
  } else {
    JS_FreeValue(ctx, JS_GetException(ctx));
  }
  JS_FreeValue(ctx, msg);
}

// Stores v as JSON as the output, and frees it. Returns 0, or 1 for an error.
static int32_t set_out_json(JSValue v) {
  JSValue json = JS_JSONStringify(ctx, v, JS_UNDEFINED, JS_UNDEFINED);
  JS_FreeValue(ctx, v);
  if (JS_IsException(json)) return fail();
  if (JS_IsUndefined(json)) {
    SET_OUT_LIT("null");
    return 0;
  }
  size_t len;
  const char *s = JS_ToCStringLen(ctx, &len, json);
  JS_FreeValue(ctx, json);
  if (!s) return fail();
  set_out(s, len);
  JS_FreeCString(ctx, s);
  return 0;
}

// The promise the code returned, while it waits for calls.
static JSValue waiting = JS_UNDEFINED;

// Runs the job queue, the code after each await and queueMicrotask
// callbacks, then stores what promise settled to, and frees it. Returns 0 for
// a result, 1 for an error and 2 if promise is still pending, which keeps it
// for kite_resolve. Jobs only run in
// ctx, the only context, so a failed job's exception is pending there. That
// includes the interrupt once the run is out of time.
static int32_t settle(JSValue promise) {
  JSContext *job_ctx;
  int job;
  while ((job = JS_ExecutePendingJob(rt, &job_ctx)) != 0) {
    if (job > 0) continue;
    // A job threw. Running out of time ends the run. Other errors, like one
    // in a queueMicrotask callback, are logged like browsers report them, as
    // nothing waits for them.
    JSValue exc = JS_GetException(ctx);
    if (JS_IsUncatchableError(exc)) {
      JS_Throw(ctx, exc);
      JS_FreeValue(ctx, promise);
      return fail();
    }
    log_error(exc);
    JS_FreeValue(ctx, exc);
  }

  JSPromiseStateEnum state = JS_PromiseState(ctx, promise);
  // The code got out of the async function Go wraps it in.
  if (state == JS_PROMISE_NOT_A_PROMISE) return set_out_json(promise);
  if (state == JS_PROMISE_PENDING) {
    waiting = promise;
    return 2;
  }
  JSValue result = JS_PromiseResult(ctx, promise);
  JS_FreeValue(ctx, promise);
  if (state == JS_PROMISE_REJECTED) {
    JS_Throw(ctx, result);
    return fail();
  }
  return set_out_json(result);
}

// Runs code, an async function expression that Go builds and null
// terminates, with input as its argument. Returns like settle.
EXPORT("kite_run")
int32_t kite_run(const char *code, uint32_t code_len, const char *input, uint32_t input_len) {
  JS_UpdateStackTop(rt);

  JSValue fn = JS_Eval(ctx, code, code_len, "code.js", JS_EVAL_TYPE_GLOBAL | JS_EVAL_FLAG_STRICT);
  if (JS_IsException(fn)) return fail();

  JSValue in = JS_ParseJSON(ctx, input, input_len, "<input>");
  if (JS_IsException(in)) {
    JS_FreeValue(ctx, fn);
    return fail();
  }

  JSValue res = JS_Call(ctx, fn, JS_UNDEFINED, 1, &in);
  JS_FreeValue(ctx, fn);
  JS_FreeValue(ctx, in);
  if (JS_IsException(res)) return fail();
  return settle(res);
}

// Settles the promise of call id with the result Go has pending, n like the
// return value of host_call, and continues the code. Returns like settle.
EXPORT("kite_resolve")
int32_t kite_resolve(uint32_t id, int32_t n) {
  JS_UpdateStackTop(rt);
  if (JS_IsUndefined(waiting) || id == 0 || id > MAX_CALLS || JS_IsUndefined(calls[id][0])) {
    SET_OUT_LIT("InternalError: unknown call");
    return 1;
  }

  JSValue res = take(ctx, n, false);
  // A failed call, or JSON QuickJS can't read.
  bool failed = JS_IsException(res);
  if (failed) res = JS_GetException(ctx);
  JSValue r = JS_Call(ctx, calls[id][failed ? 1 : 0], JS_UNDEFINED, 1, &res);
  JS_FreeValue(ctx, r);
  JS_FreeValue(ctx, res);
  JS_FreeValue(ctx, calls[id][0]);
  JS_FreeValue(ctx, calls[id][1]);
  calls[id][0] = JS_UNDEFINED;
  calls[id][1] = JS_UNDEFINED;

  JSValue p = waiting;
  waiting = JS_UNDEFINED;
  return settle(p);
}
