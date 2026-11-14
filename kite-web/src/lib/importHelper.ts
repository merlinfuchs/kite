import { helperCollectSource } from "./botghost/exporter";
import { helperReadInventorSource } from "./inventor/exporter";

// The one helper that moves commands out of the bot makers Kite can import
// from, BotGhost and inventor.bot. It's the same bookmark for both: it looks at
// which dashboard it runs on, reads the open bot there and offers what it found
// to copy. Nothing is sent anywhere: the user pastes the text into Kite
// themselves.
//
// It's plain JavaScript in a string, as it runs outside of Kite and must not
// depend on how Kite is built.

// The part that reads a dashboard. It resolves to what was found, if anything,
// and a message for the user. Kept apart from the box it's shown in, so it can
// be tested.
export const helperReadSource = String.raw`${helperCollectSource}
${helperReadInventorSource}
function findStore() {
  var all = document.querySelectorAll("body, body *");
  for (var i = 0; i < all.length; i++) {
    var keys = Object.keys(all[i]);
    for (var k = 0; k < keys.length; k++) {
      if (keys[k].indexOf("__reactContainer$") !== 0) continue;
      var stack = [all[i][keys[k]]], steps = 0;
      while (stack.length && steps++ < 100000) {
        var fiber = stack.pop();
        if (!fiber) continue;
        var store = fiber.memoizedProps && fiber.memoizedProps.store;
        if (store && typeof store.getState === "function") return store;
        if (fiber.sibling) stack.push(fiber.sibling);
        if (fiber.child) stack.push(fiber.child);
      }
    }
  }
  return null;
}
async function read() {
  var host = location.hostname;
  if (/(^|\.)botghost\.com$/.test(host)) {
    var store = findStore();
    var found = store ? collect(store.getState()) : null;
    if (!found || found.commands.length + found.events.length === 0) {
      return { message: "No commands were found. Open your bot, go to its Commands page, wait until it has loaded, and click this bookmark again." };
    }
    return {
      data: found,
      message: "Found " + found.commands.length + " commands and " + found.events.length + " custom events of this bot."
    };
  }
  if (/(^|\.)inventor\.bot$/.test(host)) {
    var data = null;
    if (/^\/dash\/[^/]+/.test(location.pathname)) {
      try { data = await readInventor(); } catch (e) { data = null; }
    }
    var flows = data ? data.items.filter(function (i) { return i.type === "flow"; }).length : 0;
    if (flows === 0) {
      return { message: "No commands were found. Open your bot in the inventor.bot dashboard and click this bookmark again." };
    }
    return { data: data, message: "Found " + flows + " commands of this bot." };
  }
  return { message: "Open your bot in the BotGhost or inventor.bot dashboard first, then click this bookmark again." };
}`;

const helperSource = String.raw`(function () {
  ${helperReadSource}
  function el(tag, style, text) {
    var e = document.createElement(tag);
    e.setAttribute("style", style);
    if (text) e.textContent = text;
    return e;
  }
  var old = document.getElementById("kite-import-helper");
  if (old) old.remove();
  var box = el("div", "position:fixed;z-index:2147483647;top:20px;right:20px;width:320px;padding:16px;border-radius:12px;background:#111827;color:#fff;font:14px/1.5 system-ui,sans-serif;box-shadow:0 10px 40px rgba(0,0,0,.5)");
  box.id = "kite-import-helper";
  box.appendChild(el("div", "font-weight:600;margin-bottom:6px", "Move to Kite"));
  var status = el("div", "color:#d1d5db;margin-bottom:12px", "Reading your bot...");
  box.appendChild(status);
  var button = "display:inline-block;margin-right:8px;padding:6px 12px;border:0;border-radius:8px;cursor:pointer;font:inherit;";
  var close = el("button", button + "background:#374151;color:#fff", "Close");
  close.onclick = function () { box.remove(); };
  box.appendChild(close);
  document.body.appendChild(box);

  read().then(function (result) {
    status.textContent = result.message;
    if (!result.data) return;
    var text = JSON.stringify(result.data);
    var copy = el("button", button + "background:#6366f1;color:#fff", "Copy for Kite");
    copy.onclick = function () {
      navigator.clipboard.writeText(text).then(function () {
        status.textContent = "Copied. Go back to Kite and paste it.";
      }, function () {
        status.textContent = "Copying was blocked. Use Save as file instead.";
      });
    };
    var save = el("button", button + "background:#374151;color:#fff", "Save as file");
    save.onclick = function () {
      var a = document.createElement("a");
      a.href = URL.createObjectURL(new Blob([text], { type: "application/json" }));
      a.download = "kite-export.json";
      a.click();
    };
    box.insertBefore(copy, close);
    box.insertBefore(save, close);
  });
})();`;

// The helper as the address of a bookmark.
export const helperBookmarklet = `javascript:${encodeURIComponent(
  helperSource
)}`;

// The helper as text to paste into the console of the browser, for browsers
// without a bookmarks bar.
export const helperConsoleSource = helperSource;
