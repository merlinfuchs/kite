import React, { useState, useMemo, useCallback, useEffect, lazy, Suspense } from 'react';
import CodeBlock from '@theme/CodeBlock';
import Box from '@mui/material/Box';
import { useColorMode } from '@docusaurus/theme-common';

const TextField = lazy(() => import('@mui/material/TextField'));

function debounce(fn, delay) {
  let timer;
  return (...args) => {
    clearTimeout(timer);
    timer = setTimeout(() => fn(...args), delay);
  };
}

export default function ParametrizedCodeBlock({ fields = [], template = '', language = 'js' }) {
  const { colorMode } = useColorMode();
  const [values, setValues] = useState(() =>
    Object.fromEntries(fields.map(({ name, default: def }) => [name, def || '']))
  );
  const [compiled, setCompiled] = useState(template);

  const compileTemplate = useCallback((vals) => {
    let result = template;

    Object.entries(vals).forEach(([key, value]) => {
      let replacement = value;
      if (key === 'field') {
        if (/\s/.test(value)) {
          replacement = `["${value}"]`;
        } else {
          replacement = `.${value}`;
        }
      }
      const re = new RegExp(`{{\\s*${key}\\s*}}`, 'g');
      result = result.replace(re, replacement);
    });

    setCompiled(result);
  }, [template]);

  const debouncedCompile = useMemo(() => debounce(compileTemplate, 300), [compileTemplate]);

  useEffect(() => {
    debouncedCompile(values);
  }, [values, debouncedCompile]);

  const handleChange = useCallback((name, value) => {
    setValues(prev => ({ ...prev, [name]: value }));
  }, []);

  return (
    <Suspense fallback={<div>Loading UI...</div>}>
      <Box sx={{ my: 2 }}>
        <Box
          sx={{
            display: 'grid',
            gap: 2,
            gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
            mb: 2,
          }}
        >
          {fields.map(({ name, label, type = 'text' }) => (
            <TextField
              key={name}
              label={label || name}
              value={values[name]}
              onChange={(e) => handleChange(name, e.target.value)}
              type={type}
              size="small"
              variant="outlined"
              placeholder={`Insert ${label || name}`}
              color={colorMode === 'dark' ? 'secondary' : 'primary'}
              InputLabelProps={{
                style: { color: colorMode === 'dark' ? '#fff' : '#000' },
              }}
              sx={{
                '& .MuiOutlinedInput-root': {
                  '& fieldset': {
                    borderColor: colorMode === 'dark' ? '#fff' : '#000',
                  },
                  '&:hover fieldset': {
                    borderColor: colorMode === 'dark' ? '#bbb' : '#555',
                  },
                  '&.Mui-focused fieldset': {
                    borderColor: '#1976d2',
                }, 
                },
                input: {
                  color: colorMode === 'dark' ? '#fff' : '#000',
                },
              }}
            />
          ))}
        </Box>
        <CodeBlock language={language}>{compiled}</CodeBlock>
      </Box>
    </Suspense>
  );
}
