// Glue between QuickJS-NG and pkg/js. Built without quickjs-libc, so scripts
// get the language and what is defined here: no files, network or timers.
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

static JSValue js_call(JSContext *ctx, JSValueConst this_val, int argc, JSValueConst *argv) {
  size_t name_len, arg_len = 4;
  const char *name = JS_ToCStringLen(ctx, &name_len, argv[0]);
  if (!name) return JS_EXCEPTION;
  JSValue arg = JS_JSONStringify(ctx, argc > 1 ? argv[1] : JS_NULL, JS_UNDEFINED, JS_UNDEFINED);
  if (JS_IsException(arg)) {
    JS_FreeCString(ctx, name);
    return arg;
  }
  const char *arg_str = "null";
  if (!JS_IsUndefined(arg)) {
    arg_str = JS_ToCStringLen(ctx, &arg_len, arg);
    if (!arg_str) {
      JS_FreeCString(ctx, name);
      JS_FreeValue(ctx, arg);
      return JS_EXCEPTION;
    }
  }

  int32_t n = host_call(name, name_len, arg_str, arg_len);
  JS_FreeCString(ctx, name);
  if (!JS_IsUndefined(arg)) JS_FreeCString(ctx, arg_str);
  JS_FreeValue(ctx, arg);

  uint32_t len = n < 0 ? -n : n;
  char *buf = js_malloc(ctx, len + 1);
  if (!buf) return JS_EXCEPTION;
  host_take(buf);
  buf[len] = 0;

  JSValue res = n < 0 ? JS_ThrowTypeError(ctx, "%s", buf) : JS_ParseJSON(ctx, buf, len, "<host>");
  js_free(ctx, buf);
  return res;
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
  JSValue stack = JS_IsError(exc) ? JS_GetPropertyStr(ctx, exc, "stack") : JS_UNDEFINED;
  size_t len;
  const char *m = JS_IsException(msg) ? NULL : JS_ToCStringLen(ctx, &len, msg);
  if (m) {
    set_out(m, len);
    JS_FreeCString(ctx, m);
  } else {
    // Like an object whose toString throws.
    SET_OUT_LIT("Error: the code threw something that can't be shown as text");
  }
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

  JSValue global = JS_GetGlobalObject(ctx);
  JSValue kite = JS_NewObject(ctx);
  JS_SetPropertyStr(ctx, kite, "call", JS_NewCFunction(ctx, js_call, "call", 2));
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

// Runs code, a function expression that Go builds and null terminates, with
// input as its argument, and stores its return value as JSON, or the error.
// A returned promise is awaited. Returns 0 for a result, 1 for an error.
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

  // Runs the code after each await and queueMicrotask callbacks. Jobs only
  // run in ctx, the only context, so a failed job's exception is pending
  // there. That includes the interrupt once the run is out of time.
  JSContext *job_ctx;
  int job;
  while ((job = JS_ExecutePendingJob(rt, &job_ctx)) > 0) {}
  if (job < 0) {
    JS_FreeValue(ctx, res);
    return fail();
  }
  if (JS_IsPromise(res)) {
    JSPromiseStateEnum state = JS_PromiseState(ctx, res);
    if (state == JS_PROMISE_PENDING) {
      JS_FreeValue(ctx, res);
      SET_OUT_LIT("the code waited for something that never finished");
      return 1;
    }
    JSValue result = JS_PromiseResult(ctx, res);
    JS_FreeValue(ctx, res);
    if (state == JS_PROMISE_REJECTED) {
      JS_Throw(ctx, result);
      return fail();
    }
    res = result;
  }

  JSValue json = JS_JSONStringify(ctx, res, JS_UNDEFINED, JS_UNDEFINED);
  JS_FreeValue(ctx, res);
  if (JS_IsException(json)) return fail();
  if (JS_IsUndefined(json)) {
    SET_OUT_LIT("null");
  } else {
    size_t len;
    const char *s = JS_ToCStringLen(ctx, &len, json);
    if (!s) {
      JS_FreeValue(ctx, json);
      return fail();
    }
    set_out(s, len);
    JS_FreeCString(ctx, s);
  }
  JS_FreeValue(ctx, json);
  return 0;
}
