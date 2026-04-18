import React from 'react';
import { render, screen } from '@testing-library/react';
import { ProfilePage } from '../../src/pages/ProfilePage';
import { AuthProvider } from '../../src/context/AuthContext';
describe('ProfilePage', () => {
  it('renders profile sections', () => {
    render(
      <AuthProvider>
        <ProfilePage />
      </AuthProvider>
    );
    // There are multiple elements with /profile/i, so use getAllByText and check at least one exists
    expect(screen.getAllByText(/profile/i).length).toBeGreaterThan(0);
    // There are multiple elements with /addresses/i, so use getAllByText and check at least one exists
    expect(screen.getAllByText(/addresses/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/contacts/i).length).toBeGreaterThan(0);
  });
});
