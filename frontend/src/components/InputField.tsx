import React from "react";

const InputField: React.FC<{ label: string; value: string | number; onChange: (value: string) => void; type: string; required?: boolean; error?: string }> = React.memo(({ label, value, onChange, type, required, error }) => (
  <div className="mb-4">
    <label className="block mb-1 text-gray-700">{label}</label>
    <input
      type={type}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className={`w-full border ${error ? "border-red-500" : "border-gray-300"} p-2 rounded focus:outline-none focus:ring-2 focus:ring-blue-500`}
      required={required}
      aria-invalid={!!error}
      aria-describedby={error ? `${label}-error` : undefined}
    />
    {error && <p id={`${label}-error`} className="text-red-500 text-sm mt-1">{error}</p>}
  </div>
));

export default InputField;