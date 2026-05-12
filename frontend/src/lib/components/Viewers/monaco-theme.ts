import type * as Monaco from "monaco-editor";

export const GRUVBOX_DARK_THEME: Monaco.editor.IStandaloneThemeData = {
  base: "vs-dark",
  inherit: true,
  rules: [
    { token: "comment", foreground: "928374", fontStyle: "italic" },
    { token: "keyword", foreground: "fb4934" },
    { token: "string", foreground: "b8bb26" },
    { token: "number", foreground: "d3869b" },
    { token: "type", foreground: "fabd2f" },
    { token: "function", foreground: "8ec07c" },
  ],
  colors: {
    "editor.background": "#1d2021",
    "editor.foreground": "#ebdbb2",
    "editor.lineHighlightBackground": "#282828",
    "editorLineNumber.foreground": "#665c54",
    "editorLineNumber.activeForeground": "#fabd2f",
    "editorCursor.foreground": "#fabd2f",
    "editor.selectionBackground": "#504945",
    "editor.inactiveSelectionBackground": "#3c3836",
    "editorIndentGuide.background": "#3c3836",
    "editorIndentGuide.activeBackground": "#504945",
    "editorGutter.background": "#1d2021",
  },
};
