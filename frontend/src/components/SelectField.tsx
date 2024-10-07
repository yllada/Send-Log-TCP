import React from "react";

const SelectField: React.FC<{ label: string; value: string; onChange: (value: string) => void; options: string[] }> = React.memo(({ label, value, onChange, options }) => (
  <div className="mb-4">
    <label className="block mb-1 text-gray-700">{label}</label>
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="w-full border border-gray-300 p-2 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
      aria-label={label}
    >
      {options.map((option) => (
        <option key={option} value={option}>{option.toUpperCase()}</option>
      ))}
    </select>
  </div>
));

export default SelectField;