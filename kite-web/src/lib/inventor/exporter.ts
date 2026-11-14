// inventor.bot has no export, but its dashboard has an API that the signed-in
// user can already read. This is the part of the shared helper (see
// importHelper.ts) that reads the components of the open bot from it. Nothing
// is sent anywhere but to inventor.bot itself, the user's own dashboard.
//
// It runs outside of Kite and must not depend on how Kite is built.
export const helperReadInventorSource = String.raw`async function readInventor() {
  var base = "/api/projects/" + location.pathname.split("/")[2] + "/components";
  var list = await (await fetch(base)).json();
  var items = [];
  for (var folder of list.folders) {
    for (var c of folder.components) {
      var blob = await (await fetch(base + "/" + c.id + "/version/" + c.current_version_id + "/blob")).json();
      items.push({ name: c.name, type: c.component_type, blob: blob });
    }
  }
  return { kite_inventor_export: 1, items: items };
}`;
