import { ReactNode } from 'react'

export const Title = ({ children }: { children: ReactNode }) => (
  <h1 className="serif text-3xl font-bold my-8">{children}</h1>
)
