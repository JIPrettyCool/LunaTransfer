"use client";
import { DatePickerInput } from '@mantine/dates';
import { useState } from 'react';
import '@mantine/dates/styles.css';

const DateSelector = () => {
  const [value, setValue] = useState<Date | null>(null);
  
  return (
    <DatePickerInput
      label="Select date"
      placeholder="Pick date"
      value={value}
      onChange={setValue}
    />
  );
};

export default DateSelector;