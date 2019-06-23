import React from 'react'
import { GradientDefs } from 'react-vis'
import Constants from '../Constants'

export const GradientDefinitions = (
  <GradientDefs>
    {Constants.colors.graphColorRange.map(color => (
      <linearGradient
        id={`bg-${color}`}
        x1="0"
        x2="0"
        y1="0"
        y2="1"
        key={color}
      >
        <stop offset="0%" stopColor={color} stopOpacity={0.5} />
        <stop offset="100%" stopColor={color} stopOpacity={0.3} />
      </linearGradient>
    ))}
  </GradientDefs>
)

export const graphColor = (i: number) =>
  Constants.colors.graphColorRange[
    i % (Constants.colors.graphColorRange.length - 1)
  ]

export const gradientDefinitionUrl = (i: number) => `url(#bg-${graphColor(i)})`
