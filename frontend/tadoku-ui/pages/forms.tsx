import { CodeBlock, Preview, Separator, Title } from '@components/example'

export default function Forms() {
  return (
    <>
      <h1 className="title mb-8">Forms</h1>

      <Title>React usage</Title>
      <Preview></Preview>
      <CodeBlock language="html" code={``} />

      <Separator />

      <Title>Basic elements</Title>
      <Preview>
        <form className="v-stack">
          <label className="label">
            <span className="label-text">First name</span>
            <input type="text" placeholder="John Doe" />
          </label>
          <label className="label">
            <span className="label-text">Message</span>
            <textarea placeholder="Dolor sit amet..." />
          </label>
          <label className="label">
            <span className="label-text">Choose a color</span>
            <select>
              <option value="#ff0000" selected>
                Red
              </option>
              <option value="#00ff00">Green</option>
              <option value="#0000ff">Blue</option>
            </select>
          </label>
          <div>
            <span className="label-text">Choose a color</span>
            <label className="label-radio">
              <input type="radio" name="color-radio" />
              <span>Red</span>
            </label>
            <label className="label-radio">
              <input type="radio" name="color-radio" />
              <span>Green</span>
            </label>
            <label className="label-radio">
              <input type="radio" name="color-radio" />
              <span>Blue</span>
            </label>
          </div>
          <label className="label error">
            <span className="label-text">First name</span>
            <input type="text" placeholder="John Doe" />
            <span className="error">Should be at least 1 character long</span>
          </label>
        </form>
      </Preview>
      <CodeBlock
        language="html"
        code={`<form className="v-stack">
<label className="label">
  <span className="label-text">First name</span>
  <input type="text" placeholder="John Doe" />
</label>
<label className="label">
  <span className="label-text">Message</span>
  <textarea placeholder="Dolor sit amet..." />
</label>
<label className="label">
  <span className="label-text">Choose a color</span>
  <select>
    <option value="#ff0000" selected>
      Red
    </option>
    <option value="#00ff00">Green</option>
    <option value="#0000ff">Blue</option>
  </select>
</label>
<div>
  <span className="label-text">Choose a color</span>
  <label className="label-radio">
    <input type="radio" name="color-radio" />
    <span>Red</span>
  </label>
  <label className="label-radio">
    <input type="radio" name="color-radio" />
    <span>Green</span>
  </label>
  <label className="label-radio">
    <input type="radio" name="color-radio" />
    <span>Blue</span>
  </label>
</div>
<label className="label error">
  <span className="label-text">First name</span>
  <input type="text" placeholder="John Doe" />
  <span className="error">Should be at least 1 character long</span>
</label>
</form>`}
      />
    </>
  )
}
