export declare class Go {
  importObject: WebAssembly.Imports;
  run(inst: WebAssembly.Instance);
}
export declare function decode(val: string, cb: (string, string) => void): void;
export declare function encode(val: string, cb: (string, string) => void): void;
