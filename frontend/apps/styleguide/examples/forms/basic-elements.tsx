export default function BasicElements() {
  return (
    <form className="v-stack spaced">
      <label className="label">
        <span className="label-text">First name</span>
        <input className="input" type="text" placeholder="John Doe" />
      </label>
      <label className="label">
        <span className="label-text">Message</span>
        <textarea className="input" placeholder="Dolor sit amet..." />
      </label>
      <label className="label">
        <span className="label-text">Choose a color</span>
        <select className="input">
          <option value="#ff0000">Red</option>
          <option value="#00ff00">Green</option>
          <option value="#0000ff">Blue</option>
        </select>
      </label>
      <div>
        <span className="label-text">Choose a color</span>
        <div className="v-stack">
          <label className="label-inline">
            <input type="radio" name="color-radio" className="input" />
            <span>Red</span>
          </label>
          <label className="label-inline">
            <input type="radio" name="color-radio" />
            <span>Green</span>
          </label>
          <label className="label-inline">
            <input type="radio" name="color-radio" />
            <span>Blue</span>
          </label>
        </div>
      </div>
      <label className="label error">
        <span className="label-text">First name</span>
        <input type="text" placeholder="John Doe" className="input" />
        <span className="error">Should be at least 1 character long</span>
      </label>
    </form>
  )
}
