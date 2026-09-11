// @vitest-environment jsdom

import React from 'react'
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { FormProvider, useForm } from 'react-hook-form'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { AmountWithUnit } from 'ui/components/Form/AmountWithUnit'

vi.mock('next/config', () => ({
  default: () => ({ publicRuntimeConfig: { apiEndpoint: '' } }),
}))

import { NewLogAPISchema } from './NewLogForm/domain'
import { NewLogV2APISchema } from './NewLogFormV2/domain'

const activity = {
  id: 1,
  name: 'Reading',
  input_type: 'amount_primary' as const,
}
const unit = {
  id: 'pages',
  unit_key: 'reading_page',
  log_activity_id: 1,
  name: 'Pages',
  modifier: 1,
}

afterEach(cleanup)

describe.each([
  ['v1', NewLogAPISchema],
  ['v2', NewLogV2APISchema],
] as const)('%s logging amount', (_, schema) => {
  function AmountForm({
    onSubmit,
    inputProps,
  }: {
    onSubmit: (payload: unknown) => void
    inputProps?: Pick<
      React.HTMLProps<HTMLInputElement>,
      'disabled' | 'readOnly' | 'onPaste'
    >
  }) {
    const methods = useForm({
      defaultValues: {
        tracking_mode: 'personal',
        registrations: [],
        selected_registrations: [],
        languageCode: 'jpn',
        activityId: 1,
        amountValue: 0,
        amountUnit: unit.id,
        allUnits: [unit],
        allActivities: [activity],
        tags: [],
      },
    })
    const preview = schema.safeParse(methods.watch())

    return (
      <FormProvider {...methods}>
        <form
          onSubmit={methods.handleSubmit(values => onSubmit(schema.parse(values)))}
        >
          <AmountWithUnit
            name="amount"
            label="Amount"
            min={0}
            step="any"
            units={[{ value: unit.id, label: unit.name }]}
            {...inputProps}
          />
          <output data-testid="preview">
            {preview.success ? JSON.stringify(preview.data) : ''}
          </output>
          <button type="submit">Create</button>
        </form>
      </FormProvider>
    )
  }

  it.each([' 12.5', '12.5 ', '\t\n12.5\r\n', '\u00a012.5\u00a0'])(
    'trims pasted %j for the live preview and submission',
    async pasted => {
      const onSubmit = vi.fn()
      render(<AmountForm onSubmit={onSubmit} />)
      const input = screen.getByLabelText('Amount', {
        selector: 'input',
      }) as HTMLInputElement

      fireEvent.paste(input, {
        clipboardData: { getData: () => pasted },
      })

      expect(input.value).toBe('12.5')
      expect(JSON.parse(screen.getByTestId('preview').textContent!)).toMatchObject({
        amount: 12.5,
      })
      fireEvent.click(screen.getByRole('button', { name: 'Create' }))
      await waitFor(() =>
        expect(onSubmit).toHaveBeenCalledWith(
          expect.objectContaining({ amount: 12.5 }),
        ),
      )
    },
  )

  it.each(['12.5', '   ', ' 12 5 ', ' 12pages ', ' 0x10 '])(
    'leaves %j to native numeric-input handling',
    pasted => {
      render(<AmountForm onSubmit={vi.fn()} />)
      const input = screen.getByLabelText('Amount', {
        selector: 'input',
      }) as HTMLInputElement

      expect(
        fireEvent.paste(input, {
          clipboardData: { getData: () => pasted },
        }),
      ).toBe(true)
      expect(input.value).toBe('0')
    },
  )

  it.each([{ disabled: true }, { readOnly: true }])(
    'does not change an input with %j',
    inputProps => {
      render(<AmountForm onSubmit={vi.fn()} inputProps={inputProps} />)
      const input = screen.getByLabelText('Amount', {
        selector: 'input',
      }) as HTMLInputElement

      fireEvent.paste(input, {
        clipboardData: { getData: () => ' 12.5 ' },
      })

      expect(input.value).toBe('0')
    },
  )

  it('honors a caller cancelling the paste', () => {
    const onPaste = vi.fn((event: React.ClipboardEvent<HTMLInputElement>) =>
      event.preventDefault(),
    )
    render(<AmountForm onSubmit={vi.fn()} inputProps={{ onPaste }} />)
    const input = screen.getByLabelText('Amount', {
      selector: 'input',
    }) as HTMLInputElement

    fireEvent.paste(input, {
      clipboardData: { getData: () => ' 12.5 ' },
    })

    expect(onPaste).toHaveBeenCalledOnce()
    expect(input.value).toBe('0')
  })
})
