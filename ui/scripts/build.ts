import { execSync } from 'child_process';
import { uiDir, getApps } from './common/utils';

const apps = getApps();

console.log(`Found apps: ${apps.join(', ')}`);

for (const app of apps) {
  console.log(`\n--- Building app: ${app} ---\n`);
  execSync('npx vite build', {
    stdio: 'inherit',
    cwd: uiDir,
    env: {
      ...process.env,
      VITE_APP_NAME: app,
    },
  });
}
