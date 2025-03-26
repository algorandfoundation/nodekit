import mocha from '@catppuccin/vscode/themes/mocha.json' with { type: 'json' };
import latte from '@catppuccin/vscode/themes/latte.json' with { type: 'json' };
import fs from 'node:fs';

/** @type {import('@astrojs/starlight/expressive-code').StarlightExpressiveCodeOptions} */
export default {
  themes: [latte, mocha],
  styleOverrides: {
    borderRadius: '0.5rem',
    borderColor: 'var(--sl-color-gray-3)',
  },
};
