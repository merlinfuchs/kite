// BotGhost has no way to export a bot, and its share codes only work inside
// BotGhost. So the user runs a helper on their own BotGhost dashboard, from a
// bookmark (see importHelper.ts, which BotGhost shares with inventor.bot). The
// part that reads the dashboard is this one: it finds the commands and custom
// events the dashboard has loaded for the open bot.
//
// It's plain ES5-style JavaScript in a string, as it runs outside of Kite and
// must not depend on how Kite is built.

// collect finds the commands and custom events in the dashboard's state,
// wherever they are in it, so the helper survives BotGhost moving them.
export const helperCollectSource = `function collect(state) {
  var commands = [], events = [], seen = [];
  function isItem(v) {
    return typeof v.name === "string" &&
      (Array.isArray(v.actions) || typeof v.compressed_tree === "string");
  }
  function visit(v, inEvents, depth) {
    if (!v || typeof v !== "object" || depth > 12 || seen.indexOf(v) !== -1) return;
    seen.push(v);
    if (Array.isArray(v)) {
      for (var i = 0; i < v.length; i++) visit(v[i], inEvents, depth + 1);
      return;
    }
    if (isItem(v)) {
      (inEvents ? events : commands).push(v);
      return;
    }
    for (var key in v) {
      if (Object.prototype.hasOwnProperty.call(v, key)) {
        visit(v[key], inEvents || /event/i.test(key), depth + 1);
      }
    }
  }
  visit(state, false, 0);
  return { kite_botghost_export: 1, commands: commands, events: events };
}`;
