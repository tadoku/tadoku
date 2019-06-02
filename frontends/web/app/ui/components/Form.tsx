import styled from 'styled-components'
import Constants from '../Constants'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { SFC, SelectHTMLAttributes } from 'react'

export const Form = styled.form``

export const Group = styled.div`
  & + & {
    margin-top: 30px;
  }
`

export const Label = styled.label`
  display: block;
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

export const LabelForRadio = styled(Label)`
  padding: 3px 8px;
  line-height: 44px;
  height: 44px;

  ${({ checked }: LabelForRadioProps) =>
    checked &&
    `
    background: ${Constants.colors.secondary};
    color: ${Constants.colors.light};
    border-radius: 3px;

    span {
      font-weight: 600;
    }
  `}

  span {
    margin-left: 8px;
  }
`
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
