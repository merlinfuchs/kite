export function initializeMonaco(monaco: any) {
  monaco.languages.typescript.typescriptDefaults.setCompilerOptions({
    target: monaco.languages.typescript.ScriptTarget.ES2016,
    allowNonTsExtensions: true,
    allowJs: true,
    moduleResolution: monaco.languages.typescript.ModuleResolutionKind.NodeJs,
    module: monaco.languages.typescript.ModuleKind.CommonJS,
    noEmit: false,
    allowImportingTsExtensions: true,
    lib: [],
  });

  monaco.languages.typescript.typescriptDefaults.setDiagnosticsOptions({
    noSemanticValidation: false,
    noSyntaxValidation: false,
  });

  const libSource = `interface Event {type: string; data: any}; interface Call {type: string; data: any;}; declare class Kite {static call(call: Call); static handle(event: Event); static describe();};`;
  const libUri = "ts:filename/global.d.ts";

  monaco.languages.typescript.typescriptDefaults.setEagerModelSync(true);
  monaco.languages.typescript.typescriptDefaults.addExtraLib(libSource, libUri);
}
