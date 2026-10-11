import { render } from '@testing-library/react';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';
import { GradientBlobs } from './GradientBlobs';

const css = readFileSync(join(process.cwd(), 'src', 'index.css'), 'utf8');

describe('GradientBlobs · capa decorativa (T1025/FR-007)', () => {
  it('son decorativos, no interceptan eventos y son dos', () => {
    const { container } = render(<GradientBlobs />);

    const root = container.firstElementChild as HTMLElement;
    expect(root).toHaveAttribute('aria-hidden', 'true');
    expect(root.className).toContain('pointer-events-none');
    expect(root.className).toContain('overflow-hidden');
    expect(root.querySelectorAll('.public-blob')).toHaveLength(2);
  });

  it('no animan con prefers-reduced-motion y solo mueven transform', () => {
    expect(css).toContain('@media (prefers-reduced-motion: reduce)');
    expect(css).toMatch(/\.public-blob\s*\{\s*animation:\s*none;/);
    // La deriva anima solo transform (compositor).
    expect(css).toMatch(/@keyframes public-marea\s*\{[\s\S]*?transform:/);
  });

  it('no aplica efectos sobre contenido de marca (no contiene imágenes)', () => {
    const { container } = render(<GradientBlobs />);

    expect(container.querySelector('img')).toBeNull();
    expect(container.querySelectorAll('[class*="filter-"]')).toHaveLength(0);
  });
});
