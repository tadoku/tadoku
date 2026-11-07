import * as stylex from "@stylexjs/stylex";
import { createElement, type ComponentPropsWithoutRef, type ReactNode } from "react";
import { sx, type StyleXProp } from "../../foundations/props";
import { link } from "../../foundations/recipes";
import {
  border,
  color,
  font,
  fontSize,
  fontWeight,
  lineHeight,
  measure,
  space,
  type,
} from "../../tokens/tokens.stylex";

const block = { marginBlockEnd: space[16], marginBlockStart: 0, marginInline: 0, maxInlineSize: measure.prose } as const;

const styles = stylex.create({
  root: { lineHeight: lineHeight.prose, overflowWrap: "break-word" },
  p: block,
  ul: { ...block, boxSizing: "border-box", paddingInlineStart: space[22] },
  ol: { ...block, boxSizing: "border-box", paddingInlineStart: space[22] },
  li: { marginBlockStart: { default: space[8], ":first-child": 0 } },
  blockquote: {
    ...block,
    borderInlineStartColor: color.ruleDefault,
    borderInlineStartStyle: "solid",
    borderInlineStartWidth: border.fieldEdge,
    boxSizing: "border-box",
    color: color.textMuted,
    paddingInlineStart: space[20],
  },
  pre: {
    ...block,
    padding: space[16],
    borderColor: color.ruleDefault,
    borderStyle: "solid",
    borderWidth: border.static,
    backgroundColor: color.surfacePaper,
    boxSizing: "border-box",
    overflowX: "auto",
  },
  h2: {
    marginInline: 0,
    fontFamily: font.display,
    fontSize: fontSize[22],
    fontWeight: fontWeight.bold,
    lineHeight: lineHeight[30],
    marginBlockEnd: space[12],
    marginBlockStart: { default: space[48], ":first-child": 0 },
    scrollMarginTop: space[24],
  },
  h3: {
    marginInline: 0,
    fontFamily: font.display,
    fontSize: type.componentSize,
    fontWeight: fontWeight.bold,
    lineHeight: lineHeight.normal,
    marginBlockEnd: space[8],
    marginBlockStart: { default: space[32], ":first-child": 0 },
    scrollMarginTop: space[24],
  },
  h4: {
    marginInline: 0,
    fontFamily: font.display,
    fontSize: fontSize[17],
    fontWeight: fontWeight.bold,
    lineHeight: lineHeight.normal,
    marginBlockEnd: space[8],
    marginBlockStart: { default: space[24], ":first-child": 0 },
    scrollMarginTop: space[24],
  },
  code: { fontFamily: font.monospace, fontSize: fontSize.em875 },
  img: { blockSize: "auto", maxInlineSize: "100%" },
  hr: {
    borderWidth: 0,
    marginBlock: space[32],
    marginInline: 0,
    borderBlockStartColor: color.ruleDefault,
    borderBlockStartStyle: "solid",
    borderBlockStartWidth: border.static,
  },
  table: {
    borderCollapse: "collapse",
    fontSize: fontSize[14],
    fontVariantNumeric: "tabular-nums",
    inlineSize: "100%",
    lineHeight: lineHeight.normal,
  },
  td: {
    paddingBlock: space[10],
    borderBlockEndColor: color.ruleDefault,
    borderBlockEndStyle: "solid",
    borderBlockEndWidth: border.static,
    paddingInlineEnd: space[16],
    paddingInlineStart: 0,
    textAlign: "left",
    verticalAlign: "top",
  },
  th: {
    paddingBlock: space[10],
    borderBlockEndColor: color.ruleStrong,
    borderBlockEndStyle: "solid",
    borderBlockEndWidth: border.static,
    color: color.textInk,
    fontWeight: fontWeight.semibold,
    paddingInlineEnd: space[16],
    paddingInlineStart: 0,
    textAlign: "left",
    verticalAlign: "top",
  },
  sup: { color: color.textLink },
});

const styled = {
  p: styles.p,
  ul: styles.ul,
  ol: styles.ol,
  li: styles.li,
  blockquote: styles.blockquote,
  pre: styles.pre,
  h2: styles.h2,
  h3: styles.h3,
  h4: styles.h4,
  code: styles.code,
  img: styles.img,
  hr: styles.hr,
  table: styles.table,
  th: styles.th,
  td: styles.td,
  sup: styles.sup,
  a: link.base,
} as const;

type StyledTag = keyof typeof styled;

type ProseElementProps<Tag extends StyledTag> = ComponentPropsWithoutRef<Tag> & {
  readonly xstyle?: StyleXProp;
  /** Supplied by react-markdown; never rendered. */
  readonly node?: unknown;
};

function element<Tag extends StyledTag>(tag: Tag) {
  function ProseElement({ className, xstyle, ...props }: ProseElementProps<Tag>) {
    const { node, ...rest } = props as typeof props & { node?: unknown };
    void node;
    return createElement(tag, { ...rest, ...sx(className, styled[tag], xstyle) });
  }
  ProseElement.displayName = `Prose.${tag}`;
  return ProseElement;
}

/**
 * Prose elements for Markdown and handwritten articles. Pass them to react-markdown as `components`, or write
 * `<Prose.p>` in JSX. Each accepts `xstyle` for article-specific layout.
 */
export const proseElements = {
  p: element("p"),
  ul: element("ul"),
  ol: element("ol"),
  li: element("li"),
  blockquote: element("blockquote"),
  pre: element("pre"),
  h2: element("h2"),
  h3: element("h3"),
  h4: element("h4"),
  code: element("code"),
  img: element("img"),
  hr: element("hr"),
  table: element("table"),
  th: element("th"),
  td: element("td"),
  sup: element("sup"),
  a: element("a"),
};

/** The prose column: reading line height and wrapping for long words. Children use `proseElements`. */
export function Prose({ children, xstyle }: { readonly children: ReactNode; readonly xstyle?: StyleXProp }) {
  return <div {...stylex.props(styles.root, xstyle)} data-prose="">{children}</div>;
}

/** A sanitized HTML fragment as plain data, so a server can parse it once and the page can render it styled. */
export type ProseNode =
  | string
  | { readonly tag: string; readonly attributes: Readonly<Record<string, string>>; readonly children: readonly ProseNode[] };

/** Elements kept from HTML; anything else is replaced by its children. */
const keptTags = new Set([
  ...Object.keys(styled),
  "h1", "h5", "h6", "strong", "b", "em", "i", "u", "s", "del", "ins", "mark", "small", "sub", "abbr", "kbd", "q",
  "br", "span", "div", "section", "figure", "figcaption", "caption", "thead", "tbody", "tfoot", "tr", "dl", "dt", "dd",
  "cite", "time",
]);
const keptAttributes = new Set([
  "href", "title", "alt", "src", "width", "height", "colspan", "rowspan", "scope", "id", "start", "reversed", "lang",
  "dir", "cite", "datetime", "rel", "target",
]);
const voidTags = new Set(["br", "hr", "img"]);

/**
 * Reads the children of a sanitized element or fragment into `ProseNode`s, dropping classes, inline styles and
 * unlisted elements. Works with any DOM implementation, such as the browser's or jsdom on a server.
 */
export function proseFromElement(parent: Node): ProseNode[] {
  const nodes: ProseNode[] = [];
  parent.childNodes.forEach((child) => {
    if (child.nodeType === 3) nodes.push(child.textContent ?? "");
    if (child.nodeType !== 1) return;
    const element = child as Element;
    const tag = element.tagName.toLowerCase();
    if (!keptTags.has(tag)) {
      nodes.push(...proseFromElement(element));
      return;
    }
    const attributes: Record<string, string> = {};
    for (const attribute of Array.from(element.attributes)) {
      if (keptAttributes.has(attribute.name)) attributes[attribute.name] = attribute.value;
    }
    nodes.push({ tag, attributes, children: voidTags.has(tag) ? [] : proseFromElement(element) });
  });
  return nodes;
}

const reactAttribute: Record<string, string> = { colspan: "colSpan", rowspan: "rowSpan", datetime: "dateTime" };

/** Renders `ProseNode`s with Paper's prose styles. */
export function ProseContent({ nodes }: { readonly nodes: readonly ProseNode[] }) {
  return <>{nodes.map(renderNode)}</>;
}

function renderNode(node: ProseNode, index: number): ReactNode {
  if (typeof node === "string") return node;
  const props: Record<string, unknown> = { key: index };
  for (const [name, value] of Object.entries(node.attributes)) props[reactAttribute[name] ?? name] = value;
  const children = node.children.length ? node.children.map(renderNode) : undefined;
  const Styled = (proseElements as Record<string, (props: Record<string, unknown>) => ReactNode>)[node.tag];
  return Styled ? createElement(Styled, props, children) : createElement(node.tag, props, children);
}
