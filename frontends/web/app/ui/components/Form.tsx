import styled from 'styled-components'
import Constants from '../../ui/Constants'

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
  font-size: 1.3em;
  margin-bottom: 7px;
`

export const LabelForRadio = styled(Label)`
  padding: 3px 0;
  line-height: 1em;

  span {
    margin-left: 5px;
  }

  input:checked + span {
    font-weight: 600;
  }
`

export const Select = styled.select`
  -moz-appearance: none;
  -webkit-appearance: none;
  appearance: none;
  border: 1px solid rgba(0, 0, 0, 0.12);
  background: ${Constants.colors.light};
  box-shadow: inset 0px 0px 2px rgba(0, 0, 0, 0.08),
    0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  padding: 4px 12px;
  font-size: 1.1em;
  height: 36px;
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
  height: 36px;
  border-radius: 3px;
  width: 100%;
  box-sizing: border-box;
`

export const Button = styled.button`
  border: none;
  background: ${Constants.colors.secondary};
  padding: 8px 20px;
  font-size: 1.1em;
  height: 36px;
  width: 100%;
  box-sizing: border-box;
`
