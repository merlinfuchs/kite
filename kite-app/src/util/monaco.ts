// @ts-ignore (we use raw-loader to get source of the declaration file)
import sdkDeclaration from "!!raw-loader!@merlingg/kite-sdk/dist/index.d.ts";

export function initializeMonaco(monaco: any) {
  monaco.languages.typescript.typescriptDefaults.setCompilerOptions({
    target: monaco.languages.typescript.ScriptTarget.ES2016,
    allowNonTsExtensions: true,
    allowJs: true,
    moduleResolution: monaco.languages.typescript.ModuleResolutionKind.NodeJs,
    module: monaco.languages.typescript.ModuleKind.CommonJS,
    noEmit: true,
    allowImportingTsExtensions: true,
    lib: [],
  });

  monaco.languages.typescript.typescriptDefaults.setDiagnosticsOptions({
    noSemanticValidation: false,
    noSyntaxValidation: false,
  });

  const warppedSDKDeclaration = `declare module "@merlingg/kite-sdk" { ${sdkDeclaration} }`;
  const sdkUri = "file:///node_modules/@merlingg/kite-sdk/index.d.ts";

  monaco.languages.typescript.typescriptDefaults.setEagerModelSync(true);
  monaco.languages.typescript.typescriptDefaults.addExtraLib(
    warppedSDKDeclaration,
    sdkUri
  );
}
