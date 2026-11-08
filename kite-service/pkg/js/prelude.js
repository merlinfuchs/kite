// Runs once before the snapshot every run starts from. Sets up the globals
// scripts get, then removes kite so scripts only use what's defined here.
(() => {
  const call = kite.call;
  const start = kite.start;
  delete globalThis.kite;

  // discord.createMessage(arg) starts a call, which Go runs while the code
  // waits for its promise. Which calls exist is up to the host of the run.
  // Names objects have anyway, like then, which await looks for, aren't
  // calls.
  for (const namespace of ["discord", "erlc", "cookieApi", "http", "variables"]) {
    const prefix = namespace + ".";
    globalThis[namespace] = new Proxy(
      {},
      {
        get: (_, name) =>
          typeof name !== "string" || name in Object.prototype || name === "then" || name === "toJSON"
            ? undefined
            : (arg) => start(prefix + name, arg),
      }
    );
  }

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
})();
