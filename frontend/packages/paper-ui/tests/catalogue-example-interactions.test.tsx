import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { phaseTwoFixtures } from '../src/catalog/phase-two';

function example(id: string) {
  const fixture = phaseTwoFixtures.find((item) => item.id === id)!;
  return render(<>{fixture.render()}</>);
}

describe('consumer examples teach observable outcomes', () => {
  it('shows the saved title after a successful Input form submission', async () => {
    const user = userEvent.setup();
    example('input.recommended');
    await user.click(screen.getByRole('button', { name: 'Save log' }));
    expect(await screen.findByRole('status')).toHaveTextContent('Saved: August reading log');
  });
  it('reports which menu action ran', async () => {
    const user = userEvent.setup();
    example('action-menu.recommended');
    await user.click(screen.getByRole('button', { name: 'Log actions' }));
    await user.click(await screen.findByRole('menuitem', { name: 'Edit log' }));
    expect(screen.getByRole('status')).toHaveTextContent('Edit log selected');
  });
  it('confirms deletion only after the modal action', async () => {
    const user = userEvent.setup();
    example('modal.recommended');
    await user.click(screen.getByRole('button', { name: 'Review deletion' }));
    await user.click(await screen.findByRole('button', { name: 'Delete log' }));
    expect(screen.getByRole('status')).toHaveTextContent('Example log deleted');
  });
});
