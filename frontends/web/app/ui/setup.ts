import { createGlobalStyle } from 'styled-components'
import Constants from './Constants'

// Setup icons
import { library } from '@fortawesome/fontawesome-svg-core'
import {
  faChevronDown,
  faEdit,
  faTrash,
  faPlus,
} from '@fortawesome/free-solid-svg-icons'
library.add(faChevronDown, faEdit, faTrash, faPlus)

// Global styles
createGlobalStyle`
  body {
    color: ${Constants.colors.dark};
  }
`
