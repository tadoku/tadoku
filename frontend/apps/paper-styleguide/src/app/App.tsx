import { catalogRegistry } from 'paper-ui/catalog'
import { Route, Routes } from 'react-router-dom'
import { DocsShell } from './DocsShell'
import { DisplayPreferencesProvider } from './preferences'
import { ResolvedCatalogueRoute } from './routes'

export function App() {
  return (
    <DisplayPreferencesProvider>
      <DocsShell documents={catalogRegistry.documents}>
        <Routes>
          <Route path="*" element={<ResolvedCatalogueRoute />} />
        </Routes>
      </DocsShell>
    </DisplayPreferencesProvider>
  )
}
