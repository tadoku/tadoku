import { Combobox, Transition } from '@headlessui/react'
import { XMarkIcon } from '@heroicons/react/20/solid'
import React, { Fragment, useState, useRef, KeyboardEvent } from 'react'
import { useController, useFormContext } from 'react-hook-form'

const MAX_TAGS = 10
const MAX_TAG_LENGTH = 50

export interface TagInputProps {
  label: string
  name: string
  hint?: string
  suggestions?: string[]
  maxTags?: number
  maxTagLength?: number
}

export function TagInput({
  label,
  name,
  hint,
  suggestions = [],
  maxTags = MAX_TAGS,
  maxTagLength = MAX_TAG_LENGTH,
}: TagInputProps) {
  const [query, setQuery] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  const { control } = useFormContext()
  const {
    field: { value: tags, onChange },
    formState: { errors },
  } = useController({ defaultValue: [], name, control })

  const currentTags: string[] = tags || []
  const hasError = errors[name] !== undefined
  const errorMessage =
    errors[name]?.message?.toString() || 'This selection is invalid'

  const canAddMore = currentTags.length < maxTags

  const normalizeTag = (tag: string): string => {
    return tag.toLowerCase().trim().slice(0, maxTagLength)
  }

  const addTag = (tag: string) => {
    const normalized = normalizeTag(tag)
    if (normalized && !currentTags.includes(normalized) && canAddMore) {
      onChange([...currentTags, normalized])
    }
    setQuery('')
  }

  const removeTag = (tagToRemove: string) => {
    onChange(currentTags.filter(tag => tag !== tagToRemove))
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault()
      if (query.trim()) {
        addTag(query)
      }
    } else if (e.key === 'Backspace' && !query && currentTags.length > 0) {
      removeTag(currentTags[currentTags.length - 1])
    }
  }

  const handleBlur = () => {
    if (query.trim()) {
      addTag(query)
    }
  }

  const filteredSuggestions =
    query === ''
      ? suggestions.filter(s => !currentTags.includes(s.toLowerCase()))
      : suggestions.filter(
          s =>
            s.toLowerCase().includes(query.toLowerCase()) &&
            !currentTags.includes(s.toLowerCase()),
        )

  const id = `${name}-tag-input`

  return (
    <label className={`label ${hasError ? 'error' : ''}`} htmlFor={id}>
      <span className="label-text">
        {label}
        {hint ? (
          <span className="label-hint hidden sm:flex">{hint}</span>
        ) : undefined}
      </span>
      <Combobox
        value={null}
        onChange={(selected: string | null) => {
          if (selected) {
            addTag(selected)
          }
        }}
      >
        <div className="input relative min-h-[42px] flex flex-wrap gap-1 items-center py-1">
          {currentTags.map(tag => (
            <span
              key={tag}
              className="inline-flex items-center gap-1 px-2 py-0.5 bg-secondary/10 text-secondary rounded text-sm"
            >
              {tag}
              <button
                type="button"
                onClick={e => {
                  e.preventDefault()
                  removeTag(tag)
                }}
                className="hover:bg-secondary/20 rounded"
              >
                <XMarkIcon className="h-4 w-4" />
              </button>
            </span>
          ))}
          {canAddMore && (
            <Combobox.Input
              ref={inputRef}
              id={id}
              value={query}
              onChange={e => setQuery(e.target.value)}
              onKeyDown={handleKeyDown}
              onBlur={handleBlur}
              placeholder={currentTags.length === 0 ? 'Type a tag...' : ''}
              className="flex-1 min-w-[100px] border-none outline-none bg-transparent p-0 text-sm"
              autoComplete="off"
            />
          )}
          <Transition
            as={Fragment}
            leave="transition ease-in duration-100"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
            afterLeave={() => {}}
          >
            <Combobox.Options className="absolute left-0 top-full mt-1 z-50 max-h-60 w-full overflow-auto bg-white py-1 shadow-md shadow-slate-500/20 ring-1 ring-secondary ring-opacity-5 focus:outline-none">
              {query.trim() &&
                !filteredSuggestions
                  .map(s => s.toLowerCase())
                  .includes(query.toLowerCase()) && (
                  <Combobox.Option
                    value={query}
                    className={({ active }) =>
                      `relative cursor-pointer select-none py-2 px-4 ${
                        active ? 'bg-secondary text-white' : ''
                      }`
                    }
                  >
                    Create "{normalizeTag(query)}"
                  </Combobox.Option>
                )}
              {filteredSuggestions.slice(0, 10).map(suggestion => (
                <Combobox.Option
                  key={suggestion}
                  value={suggestion}
                  className={({ active }) =>
                    `relative cursor-pointer select-none py-2 px-4 ${
                      active ? 'bg-secondary text-white' : ''
                    }`
                  }
                >
                  {suggestion}
                </Combobox.Option>
              ))}
              {filteredSuggestions.length === 0 && !query.trim() && (
                <div className="relative cursor-default select-none py-2 px-4 text-gray-500">
                  Type to add a tag
                </div>
              )}
            </Combobox.Options>
          </Transition>
        </div>
      </Combobox>
      <div className="flex justify-between">
        <span className="text-xs text-gray-500">
          {currentTags.length}/{maxTags} tags
        </span>
        {query.length > 0 && (
          <span className="text-xs text-gray-500">
            {query.length}/{maxTagLength} chars
          </span>
        )}
      </div>
      {hint ? <span className="label-hint sm:hidden">{hint}</span> : undefined}
      <span className="error">{errorMessage}</span>
    </label>
  )
}
