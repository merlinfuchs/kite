// Runs once before the snapshot every run starts from. Sets up the globals
// scripts get, then removes kite so scripts only use what's defined here.
(() => {
  const call = kite.call;
  const startCall = kite.start;
  delete globalThis.kite;

  // While the top level of a module runs, what it registered, see
  // __kiteModule below. Calls fail then, and the first one that did is kept,
  // so the module fails to load even if the code catches it or doesn't wait
  // for it.
  let loading = null;
  let loadError = null;
  const noCallsWhileLoading = (name) => {
    if (!loading) return;
    const err = new Error(`${name} can't be used at the top level of a module. Call it inside a handler instead.`);
    loadError ??= err;
    throw err;
  };
  const start = (name, arg) => {
    noCallsWhileLoading(name);
    return startCall(name, arg);
  };

  // discord.createMessage(arg) starts a call, which Go runs while the code
  // waits for its promise. Which calls exist is up to the host of the run.
  // Names objects have anyway, like then, which await looks for, aren't
  // calls. Names in registers are the functions that register handlers of
  // modules.
  const namespaces = {};
  const registers = {};
  for (const namespace of ["discord", "erlc", "cookieApi", "http", "variables", "ai"]) {
    const prefix = namespace + ".";
    const own = (registers[namespace] = {});
    globalThis[namespace] = namespaces[namespace] = new Proxy(
      {},
      {
        get: (_, name) =>
          typeof name !== "string" || name in Object.prototype || name === "then" || name === "toJSON"
            ? undefined
            : Object.hasOwn(own, name)
              ? own[name]
              : (arg) => start(prefix + name, arg),
      }
    );
  }

  // fetch(url, init) sends the request with http.request, for code written
  // for fetch. It takes a text body and headers as an object or pairs, and
  // the response has ok, status, statusText, headers.get(), text() and
  // json(). Like fetch, only failing to get a response rejects.
  globalThis.fetch = async (url, init = {}) => {
    // Before the call, so the error names fetch.
    noCallsWhileLoading("fetch");
    const { method, headers = {}, body } = init ?? {};
    if (body != null && typeof body !== "string") {
      throw new TypeError("fetch: the body must be text, like JSON.stringify(data)");
    }
    const pairs = Array.isArray(headers) ? headers : Object.entries(headers);
    const res = await namespaces.http.request({
      url: String(url),
      method: method == null ? undefined : String(method).toUpperCase(),
      headers: Object.fromEntries(pairs.map(([name, value]) => [String(name), String(value)])),
      body: body ?? undefined,
    });
    const header = (name) => {
      const key = String(name).toLowerCase();
      return Object.hasOwn(res.headers, key) ? res.headers[key] : null;
    };
    return {
      ok: res.status >= 200 && res.status < 300,
      status: res.status,
      statusText: res.status_text,
      url: String(url),
      headers: { get: header, has: (name) => header(name) !== null },
      text: async () => res.body,
      json: async () => JSON.parse(res.body),
    };
  };

  const show = (v) => {
    if (typeof v === "string") return v;
    if (v instanceof Error) return String(v);
    if (v === undefined) return "undefined";
    try {
      return JSON.stringify(v) ?? String(v);
    } catch {
      return String(v);
    }
  };
  for (const name of ["log", "info", "warn", "error"]) {
    const log = console[name];
    console[name] = (...args) => log(args.map(show).join(" "));
  }

  const months = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];
  const days = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];
  const pad = (n, len = 2) => String(n).padStart(len, "0");

  // formatDate(date, format, timeZone) formats a Date, a timestamp in
  // milliseconds or a date string, like formatDate(Date.now(), "dddd HH:mm",
  // "Europe/Berlin"). Text in [brackets] is kept as it is.
  globalThis.formatDate = (date = Date.now(), format = "YYYY-MM-DD HH:mm", timeZone = "UTC") => {
    const time = date instanceof Date ? date.getTime() : typeof date === "string" ? Date.parse(date) : Number(date);
    if (!Number.isFinite(time)) throw new TypeError("formatDate: invalid date " + show(date));
    const p = call("date_parts", { time, time_zone: String(timeZone) });
    const h12 = p.hour % 12 || 12;
    const offset = (p.offset < 0 ? "-" : "+") + pad(Math.floor(Math.abs(p.offset) / 60)) + ":" + pad(Math.abs(p.offset) % 60);
    const tokens = {
      YYYY: pad(p.year, 4), YY: pad(p.year % 100), MMMM: months[p.month - 1], MMM: months[p.month - 1].slice(0, 3),
      MM: pad(p.month), M: p.month, DD: pad(p.day), D: p.day, dddd: days[p.weekday], ddd: days[p.weekday].slice(0, 3),
      HH: pad(p.hour), H: p.hour, hh: pad(h12), h: h12, mm: pad(p.minute), m: p.minute, ss: pad(p.second), s: p.second,
      A: p.hour < 12 ? "AM" : "PM", a: p.hour < 12 ? "am" : "pm", Z: offset,
    };
    return String(format).replace(/\[([^\]]*)]|YYYY|YY|MMMM|MMM|MM|M|DD|D|dddd|ddd|HH|H|hh|h|mm|m|ss|s|A|a|Z/g, (t, literal) => literal ?? String(tokens[t]));
  };

  // There's no Intl, so toLocaleString on numbers gives 1234.5. Group the
  // digits like en-US instead, which is what most code expects.
  const group = (s) => s.replace(/\B(?=(\d{3})+(?!\d))/g, ",");
  Object.defineProperty(Number.prototype, "toLocaleString", {
    value: function (locales, options = {}) {
      const n = Number(this);
      if (!Number.isFinite(n)) return String(n);
      const min = options.minimumFractionDigits ?? 0;
      const max = Math.max(min, options.maximumFractionDigits ?? 3);
      let [int, frac = ""] = Math.abs(n).toFixed(max).split(".");
      frac = frac.replace(/0+$/, "").padEnd(min, "0");
      return (n < 0 ? "-" : "") + group(int) + (frac ? "." + frac : "");
    },
    writable: true,
    configurable: true,
  });

  // Modules register their handlers at the top level, with discord.command,
  // discord.on and discord.component, and declare their stored variables
  // with variables.define. Each of them throws outside the top level, as
  // what it registers only counts while the module loads.
  const registering = (name) => {
    if (!loading) throw new Error(`${name} can only be used at the top level of a module`);
    return loading;
  };
  const isObject = (v) => v !== null && typeof v === "object" && !Array.isArray(v);
  const checkHandler = (fn, handler) => {
    if (typeof handler !== "function") throw new TypeError(`${fn}: the handler must be a function, like async (ctx) => { ... }`);
  };

  const isCommandName = (name) =>
    typeof name === "string" && name.split(" ").length <= 3 && name.split(" ").every((word) => /^[-_'\p{Ll}\p{Lo}\p{N}]{1,32}$/u.test(word));

  // discord.command("level show", { description, options }, handler)
  // registers a command, or a subcommand with the names of its command and
  // group before it. The options are Discord's, the handler gets ctx with
  // the interaction.
  registers.discord.command = (name, options, handler) => {
    const r = registering("discord.command");
    if (typeof options === "function") [options, handler] = [{}, options];
    if (!isObject(options)) throw new TypeError('discord.command: the options must be an object, like { description: "Shows your level" }');
    // User and message commands, which Discord shows in menus, have names
    // like "Show level". Slash commands have lowercase words for the
    // command, group and subcommand.
    const menu = options.type === 2 || options.type === 3;
    if (menu ? typeof name !== "string" || name.length < 1 || name.length > 32 : !isCommandName(name)) {
      throw new TypeError(
        menu
          ? "discord.command: the name of a user or message command must be 1 to 32 characters"
          : 'discord.command: the name must be up to 3 lowercase words of up to 32 letters, digits, - and _, like "level show"'
      );
    }
    checkHandler("discord.command", handler);
    if (r.commands.has(name)) throw new TypeError(`discord.command: there's already a command called "${name}"`);
    if (!menu) {
      for (const [other, { data }] of r.commands) {
        if (data.type === 2 || data.type === 3) continue;
        const [parent, child] = other.length < name.length ? [other, name] : [name, other];
        if (child.startsWith(parent + " ")) {
          throw new TypeError(`discord.command: "${parent}" can't be a command and also have subcommands like "${child}"`);
        }
      }
    }
    // A copy, as the module could change the object later, and to fail here
    // for options that aren't JSON.
    r.commands.set(name, { data: JSON.parse(JSON.stringify(options)), handler });
  };

  // discord.on("message_create", handler) registers a handler of an event,
  // by the name of Discord's gateway event in lower case. An event can have
  // more than one, which run one after another.
  registers.discord.on = (event, handler) => {
    const r = registering("discord.on");
    if (typeof event !== "string" || !/^[a-z_]+$/.test(event)) throw new TypeError('discord.on: the event must be a name like "message_create"');
    checkHandler("discord.on", handler);
    if (!r.events.has(event)) r.events.set(event, []);
    r.events.get(event).push(handler);
  };

  // discord.component("claim", handler) registers a handler of buttons,
  // select menus and modals, which they name when the module sends them.
  registers.discord.component = (name, handler) => {
    const r = registering("discord.component");
    if (typeof name !== "string" || !/^[\w-]{1,32}$/.test(name)) {
      throw new TypeError("discord.component: the name must be up to 32 letters, digits, - and _");
    }
    checkHandler("discord.component", handler);
    if (r.components.has(name)) throw new TypeError(`discord.component: there's already a component called "${name}"`);
    r.components.set(name, handler);
  };

  // variables.define("xp", { scoped: true }) declares a stored variable the
  // module uses by its name, with one value per scope if scoped.
  registers.variables.define = (name, options = {}) => {
    const r = registering("variables.define");
    if (typeof name !== "string" || name === "") throw new TypeError('variables.define: the name must be text, like "xp"');
    if (!isObject(options)) throw new TypeError("variables.define: the options must be an object, like { scoped: true }");
    const scoped = options.scoped ?? false;
    if (typeof scoped !== "boolean") throw new TypeError("variables.define: scoped must be true or false");
    if (r.variables.has(name)) throw new TypeError(`variables.define: there's already a variable called "${name}"`);
    r.variables.set(name, { scoped });
  };

  // The values of the options of a command by their names, from the
  // options of the interaction, where subcommands and groups nest theirs.
  const optionValues = (options = [], values = {}) => {
    for (const option of options) {
      if (option.type === 1 || option.type === 2) {
        optionValues(option.options, values);
      } else {
        values[option.name] = option.value;
      }
    }
    return values;
  };

  // The responses to an interaction, Discord's callback types. Go answers
  // the interaction of the run, and the first response has to come within 3
  // seconds.
  const respond = (type) => (data) => start("interaction.respond", data === undefined ? { type } : { type, data });
  const interactionContext = (interaction) => ({
    interaction,
    reply: respond(4),
    defer: respond(5),
    deferUpdate: respond(6),
    update: respond(7),
    showModal: respond(9),
    followUp: (data) => start("interaction.followUp", data),
  });

  const manifest = (r) => ({
    commands: [...r.commands].map(([name, { data }]) => ({ name, data })),
    events: [...r.events.keys()],
    components: [...r.components.keys()],
    variables: [...r.variables].map(([name, { scoped }]) => ({ name, scoped })),
  });

  // Go loads a module by running __kiteModule(function () { <module> }),
  // which returns the function kite_run calls. That runs the top level of the
  // module, keeps what it registered and returns it. Go takes a snapshot of
  // the instance then, and every run of a handler starts from it with
  // __kiteDispatch, which kite_run calls with the command, event or component
  // to handle.
  let loaded = null;
  const loadModule = (body) => async () => {
    const r = { commands: new Map(), events: new Map(), components: new Map(), variables: new Map() };
    loading = r;
    loadError = null;
    try {
      body();
    } finally {
      loading = null;
    }
    if (loadError) throw loadError;
    loaded = r;
    return manifest(r);
  };

  const dispatch = async (input) => {
    const r = loaded;
    if (!r) throw new Error("the module isn't loaded");
    switch (input.kind) {
      case "command":
      case "component": {
        const command = input.kind === "command";
        const handler = command ? r.commands.get(input.name)?.handler : r.components.get(input.name);
        if (!handler) throw new Error(`the module has no ${input.kind} called "${input.name}"`);
        const ctx = interactionContext(input.interaction);
        if (command) {
          ctx.options = optionValues(input.interaction?.data?.options);
        } else {
          ctx.data = input.data ?? null;
        }
        await handler(ctx);
        break;
      }
      case "event": {
        const handlers = r.events.get(input.name);
        if (!handlers) throw new Error(`the module has no handler for the event ${input.name}`);
        // A failing handler doesn't stop the others. The first error is the
        // error of the run, the others are logged.
        const errors = [];
        for (const handler of handlers) {
          try {
            await handler({ event: input.event });
          } catch (err) {
            errors.push(err);
          }
        }
        for (const err of errors.slice(1)) console.error(show(err));
        if (errors.length > 0) throw errors[0];
        break;
      }
      default:
        throw new Error(`unknown kind ${input.kind}`);
    }
    return null;
  };
  Object.defineProperty(globalThis, "__kiteModule", { value: loadModule });
  Object.defineProperty(globalThis, "__kiteDispatch", { value: dispatch });
})();
