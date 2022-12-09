import { withErrorBoundary, FallbackProps } from 'react-error-boundary'

export const ErrorFallback = ({ error, resetErrorBoundary }: FallbackProps) => (
  <>
    <h1>Something seems to have gone wrong</h1>
    <p>
      <a onClick={resetErrorBoundary} href="#">
        Please refresh this page to try again.
      </a>
    </p>
    <pre>{error.message}</pre>
  </>
)

export const withOryErrorBoundary = <P extends object>(
  ComponentThatMayError: React.ComponentType<P>,
) =>
  withErrorBoundary(ComponentThatMayError, {
    FallbackComponent: ErrorFallback,
    onError(error, info) {
      // Do something with the error
      // E.g. log to an error logging client here
    },
  })
