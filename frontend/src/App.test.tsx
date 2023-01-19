import React from 'react';
import { render, screen } from '@testing-library/react';
import App from './App';

test('renders eCommerce title', () => {
  render(<App />);
  const linkElement = screen.getByText(/eCommerce/i);
  expect(linkElement).toBeInTheDocument();
});
