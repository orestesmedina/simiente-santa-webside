import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { App } from './App';

describe('App', () => {
  it('renderiza la aplicación en su estado inicial', () => {
    render(<App />);

    expect(screen.getByRole('heading', { name: 'Estado del sistema' })).toBeInTheDocument();
  });
});
