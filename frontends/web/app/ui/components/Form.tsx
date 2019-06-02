import styled from 'styled-components'
import Constants from '../Constants'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { SFC, SelectHTMLAttributes, InputHTMLAttributes } from 'react'

export const Form = styled.form``

export const Group = styled.div`
  & + & {
    margin-top: 30px;
  }
`

export const Label = styled.label`
  display: block;

  & + & {
    margin-top: 12px;
  }
`
export const LabelText = styled.span`
  display: block;
  font-weight: 600;
  font-size: 1.1em;
  margin-bottom: 7px;
`

interface LabelForRadioProps {
  checked?: boolean
}

const RadioLabel = styled(Label)`
  display: flex;
  align-items: center;
  padding: 3px 8px;
  height: 44px;
  transition: all 0.2s ease;

  ${({ checked }: LabelForRadioProps) =>
    checked &&
    `
    background: ${Constants.colors.lightGray};
    border-radius: 3px;
    box-shadow: inset 0px 0px 2px rgba(0, 0, 0, 0.08),
      0px 2px 3px 0px rgba(0, 0, 0, 0.08);

    span {
      font-weight: 600;
    }
  `}

  span {
    margin-left: 8px;
  }
`

const HiddenRadio = styled.input`
  display: none;
`

const StyledRadio = styled.span`
  display: inline-block;
  height: 10px;
  width: 10px;
  border-radius: 8px;
  border: 2px solid ${Constants.colors.secondary};
  transition: all 0.2s ease;

  ${({ checked }: LabelForRadioProps) =>
    checked &&
    `
    background:  ${Constants.colors.primary};
    border: 2px solid white;
  `}
`

export const RadioButton: SFC<
  InputHTMLAttributes<HTMLInputElement> & {
    label: string
  }
> = ({ label, checked, ...props }) => (
  <RadioLabel checked={checked}>
    <HiddenRadio type="radio" {...props} checked={checked} />
    <StyledRadio checked={checked} />
    <span>{label}</span>
  </RadioLabel>
)

export const SelectGroup = styled.div`
  position: relative;
`

const SelectArrow = styled(FontAwesomeIcon)`
  font-size: 12px;
  position: absolute;
  top: 16px;
  right: 16px;
  color: #434b67;
  pointer-events: none;
`

export const Select: SFC<SelectHTMLAttributes<HTMLSelectElement>> = ({
  children,
  ...props
}) => (
  <SelectGroup>
    <SelectArrow icon="chevron-down" />
    <StyledSelect {...props}>{children}</StyledSelect>
  </SelectGroup>
)

const StyledSelect = styled.select`
  -moz-appearance: none;
  -webkit-appearance: none;
  appearance: none;
  border: 1px solid rgba(0, 0, 0, 0.12);
  background: ${Constants.colors.light};
  box-shadow: inset 0px 0px 2px rgba(0, 0, 0, 0.08),
    0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  padding: 4px 20px 4px 12px;
  font-size: 1.1em;
  height: 44px;
  border-radius: 3px;
  width: 100%;
  box-sizing: border-box;
`

export const Input = styled.input`
  border: 1px solid rgba(0, 0, 0, 0.12);
  background: ${Constants.colors.light};
  box-shadow: inset 0px 0px 2px rgba(0, 0, 0, 0.08),
    0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  padding: 4px 12px;
  font-size: 1.1em;
  height: 44px;
  border-radius: 3px;
  width: 100%;
  box-sizing: border-box;
`
