import type { McpUiStyleVariableKey } from '@modelcontextprotocol/ext-apps';

export const getCssVar = (variable: McpUiStyleVariableKey) => `var(${variable})`;
