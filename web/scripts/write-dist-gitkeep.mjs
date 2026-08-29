// Recreates dist/.gitkeep after every build.
//
// Vite empties outDir, which deletes the placeholder. The file has to exist in a
// clean checkout because the Go binary embeds web/dist, and an embed directive
// with no matching files fails the build — so without it the backend stops
// compiling for anyone who has not run the frontend build yet.
import { closeSync, mkdirSync, openSync } from "node:fs";

mkdirSync("dist", { recursive: true });
closeSync(openSync("dist/.gitkeep", "a"));
