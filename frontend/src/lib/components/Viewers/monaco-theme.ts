import type * as Monaco from "monaco-editor";

// Monaco는 CSS 변수를 못 읽어서 app.css Signal 토큰 값을 16진수로 옮겨 둠
export const SIGNAL_THEME: Monaco.editor.IStandaloneThemeData = {
  base: "vs-dark",
  inherit: true,
  rules: [
    { token: "comment", foreground: "a1a5ab", fontStyle: "italic" },
    { token: "keyword", foreground: "f19ac4" },
    { token: "string", foreground: "8fd9a8" },
    { token: "number", foreground: "9cc8f5" },
    { token: "type", foreground: "f3c27a" },
    { token: "function", foreground: "86d8e3" },
  ],
  colors: {
    "editor.background": "#090a0d",
    "editor.foreground": "#edeef1",
    "editor.lineHighlightBackground": "#131518",
    "editorLineNumber.foreground": "#71757a",
    "editorLineNumber.activeForeground": "#edeef1",
    "editorCursor.foreground": "#bdf952",
    "editor.selectionBackground": "#bdf95240",
    "editor.inactiveSelectionBackground": "#bdf95220",
    "editorIndentGuide.background": "#2b2e32",
    "editorIndentGuide.activeBackground": "#71757a",
    "editorGutter.background": "#090a0d",
    focusBorder: "#bdf952",
  },
};
