import {
  Combobox,
  ComboboxInput,
  ComboboxOptions,
  ComboboxOption,
} from '@headlessui/react'
import { XMarkIcon } from '@heroicons/react/20/solid'
import React, { useState } from 'react'
import { useController, useFormContext } from 'react-hook-form'

export function TagsInput(props: {
  label: string
  name: string
  hint?: string
  getSuggestions: (inputText: string) => string[]
  placeholder?: string
}) {
  const [query, setQuery] = useState('')

  const { name, label, hint, getSuggestions, placeholder } = props
  const { control } = useFormContext()
  const {
    field: { value, onChange },
    formState: { errors },
  } = useController({ defaultValue: [], name, control })

  const tags: string[] = value || []
  const hasError = errors[name] !== undefined
  const errorMessage =
    errors[name]?.message?.toString() || 'This selection is invalid'

  const suggestions = getSuggestions(query).filter(s => !tags.includes(s))

  const handleSelect = (selected: string | null) => {
    if (selected && !tags.includes(selected)) {
      onChange([...tags, selected])
    }
    setQuery('')
  }

  const handleRemove = (tagToRemove: string) => {
    onChange(tags.filter(tag => tag !== tagToRemove))
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && query.trim() && suggestions.length === 0) {
      e.preventDefault()
      if (!tags.includes(query.trim())) {
        onChange([...tags, query.trim()])
      }
      setQuery('')
    }
  }

  // Needs to be suffixed with -search so 1password doesn't try to autocomplete...
  const id = `${name}-search`

  return (
    <div className={`label ${hasError ? 'error' : ''}`}>
      <span className="label-text">
        {label}
        {hint ? (
          <span className="label-hint hidden sm:flex">{hint}</span>
        ) : undefined}
      </span>
      {tags.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {tags.map(tag => (
            <span
              key={tag}
              className="tag bg-secondary/20 text-secondary-900 rounded-md"
            >
              {tag}
              <button
                type="button"
                onClick={() => handleRemove(tag)}
                className="ml-1 hover:text-red-600 focus:outline-none"
                aria-label={`Remove ${tag}`}
              >
                <XMarkIcon className="h-4 w-4" />
              </button>
            </span>
          ))}
        </div>
      )}
      <Combobox
        value={null}
        onChange={handleSelect}
        onClose={() => setQuery('')}
      >
        <div className="input relative">
          <ComboboxInput
            id={id}
            value={query}
            onChange={event => setQuery(event.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            className="w-full"
          />
          <ComboboxOptions
            transition
            className={`absolute mt-2 z-50 max-h-60 w-full overflow-auto bg-white py-1 shadow-md shadow-slate-500/20 ring-1 ring-secondary ring-opacity-5 focus:outline-none transition ease-in duration-100 data-[closed]:opacity-0`}
          >
            {suggestions.length === 0 && query !== '' ? (
              <div className="relative cursor-default select-none py-2 px-4 text-gray-700">
                Press Enter to add &quot;{query}&quot;
              </div>
            ) : (
              suggestions.map(suggestion => (
                <ComboboxOption
                  key={suggestion}
                  value={suggestion}
                  className="relative cursor-default select-none py-2 px-4 data-[focus]:bg-secondary data-[focus]:text-white"
                >
                  <span className="block truncate">{suggestion}</span>
                </ComboboxOption>
              ))
            )}
          </ComboboxOptions>
        </div>
      </Combobox>
      {hint ? <span className="label-hint sm:hidden">{hint}</span> : undefined}
      <span className="error">{errorMessage}</span>
    </div>
  )
}
