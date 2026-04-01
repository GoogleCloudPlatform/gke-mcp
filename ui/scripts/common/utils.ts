import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
export const uiDir = path.join(__dirname, '..', '..');
export const appsDir = path.join(uiDir, 'apps');

export function getApps(): string[] {
  return fs.readdirSync(appsDir).filter((file: string) => {
    if (file === '.' || file === '..') {
      return false;
    }
    try {
      return fs.statSync(path.join(appsDir, file)).isDirectory();
    } catch {
      return false;
    }
  });
}
