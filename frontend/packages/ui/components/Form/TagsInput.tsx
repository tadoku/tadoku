import {
  Combobox,
  ComboboxInput,
  ComboboxOptions,
  ComboboxOption,
} from '@headlessui/react'
import { XMarkIcon } from '@heroicons/react/20/solid'
import React, { useState, useEffect, useRef } from 'react'
import { useController, useFormContext } from 'react-hook-form'

export function TagsInput(props: {
  label: string
  name: string
  hint?: string
  getSuggestions: (inputText: string) => string[] | Promise<string[]>
  placeholder?: string
  debounceMs?: number
  maxTags?: number
}) {
  const [query, setQuery] = useState('')
  const [suggestions, setSuggestions] = useState<string[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const {
    name,
    label,
    hint,
    getSuggestions,
    placeholder,
    debounceMs = 300,
    maxTags,
  } = props
  const { control } = useFormContext()
  const {
    field: { value, onChange },
    formState: { errors },
  } = useController({ defaultValue: [], name, control })

  const tags: string[] = value || []
  const hasError = errors[name] !== undefined
  const errorMessage =
    errors[name]?.message?.toString() || 'This selection is invalid'

  useEffect(() => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current)
    }

    setIsLoading(true)

    debounceRef.current = setTimeout(async () => {
      try {
        const result = await Promise.resolve(getSuggestions(query))
        const filtered = result.filter(s => !tags.includes(s))
        setSuggestions(filtered)
      } catch {
        setSuggestions([])
      } finally {
        setIsLoading(false)
      }
    }, debounceMs)

    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current)
      }
    }
  }, [query, tags, getSuggestions, debounceMs])

  const isAtLimit = maxTags !== undefined && tags.length >= maxTags

  const handleSelect = (selected: string | null) => {
    if (selected && !tags.includes(selected) && !isAtLimit) {
      onChange([...tags, selected])
    }
    setQuery('')
  }

  const handleRemove = (tagToRemove: string) => {
    onChange(tags.filter(tag => tag !== tagToRemove))
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && query.trim() && suggestions.length === 0 && !isLoading) {
      e.preventDefault()
      if (!tags.includes(query.trim()) && !isAtLimit) {
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
        {maxTags !== undefined ? (
          <span className="label-hint hidden sm:flex">{tags.length}/{maxTags}</span>
        ) : hint ? (
          <span className="label-hint hidden sm:flex">{hint}</span>
        ) : undefined}
      </span>
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
            placeholder={isAtLimit ? 'Maximum tags reached' : placeholder}
            disabled={isAtLimit}
            className="w-full"
          />
          <ComboboxOptions
            modal={false}
            transition
            className={`absolute mt-2 z-50 max-h-60 w-full overflow-auto bg-white py-1 shadow-md shadow-slate-500/20 ring-1 ring-secondary ring-opacity-5 focus:outline-none transition ease-in duration-100 data-[closed]:opacity-0`}
          >
            {isLoading ? (
              <div className="relative cursor-default select-none py-2 px-4 text-gray-500">
                Loading...
              </div>
            ) : suggestions.length === 0 && query !== '' ? (
              <div className="relative cursor-default select-none py-2 px-4 text-gray-700">
                Press Enter to add &quot;{query}&quot;
              </div>
            ) : suggestions.length === 0 ? (
              <div className="relative cursor-default select-none py-2 px-4 text-gray-500">
                No suggestions available
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
      {tags.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {tags.map(tag => (
            <span
              key={tag}
              className="tag bg-black/5 text-secondary-900 pr-0 border border-b-2 border-black/5"
            >
              {tag}
              <button
                type="button"
                onClick={() => handleRemove(tag)}
                className="hover:bg-black/5 focus:outline-none px-2 -my-2 py-2 ml-1"
                aria-label={`Remove ${tag}`}
              >
                <XMarkIcon className="h-5 w-5" />
              </button>
            </span>
          ))}
        </div>
      )}
      {maxTags !== undefined ? (
        <span className="label-hint sm:hidden">{tags.length}/{maxTags}</span>
      ) : hint ? (
        <span className="label-hint sm:hidden">{hint}</span>
      ) : undefined}
      <span className="error">{errorMessage}</span>
    </div>
  )
}
