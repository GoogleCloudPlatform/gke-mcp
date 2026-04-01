import { execSync } from 'child_process';
import { uiDir, getApps } from './common/utils';

const apps = getApps();

console.log(`Found apps: ${apps.join(', ')}`);

let failed = false;

for (const app of apps) {
  console.log(`\n--- Testing app: ${app} ---\n`);
  try {
    execSync(`npx vitest run apps/${app} --passWithNoTests`, {
      stdio: 'inherit',
      cwd: uiDir,
      env: {
        ...process.env,
        VITE_APP_NAME: app,
      },
    });
  } catch {
    console.error(`\n❌ Tests failed for app: ${app}\n`);
    failed = true;
  }
}

if (failed) {
  console.error('\n❌ UI tests failed for one or more apps.\n');
  process.exit(1);
} else {
  console.log('\n✅ All UI tests passed.\n');
}
