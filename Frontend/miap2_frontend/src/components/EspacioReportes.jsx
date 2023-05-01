import * as React from 'react';
import PropTypes from 'prop-types';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Service from '../Services/Service';
import { useState, useEffect } from 'react';
import { Graphviz } from 'graphviz-react';

function TabPanel(props) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`vertical-tabpanel-${index}`}
      aria-labelledby={`vertical-tab-${index}`}
      {...other}
    >
      {value === index && (
        <Box sx={{ p: 3 }}>
          <Typography>{children}</Typography>
        </Box>
      )}
    </div>
  );
}

TabPanel.propTypes = {
  children: PropTypes.node,
  index: PropTypes.number.isRequired,
  value: PropTypes.number.isRequired,
};

function a11yProps(index) {
  return {
    id: `vertical-tab-${index}`,
    'aria-controls': `vertical-tabpanel-${index}`,
  };
}

export default function VerticalTabs() {
  const [value, setValue] = React.useState(0);
  const [tabs, setTabs] = useState([]);

  const handleChange = (event, newValue) => {
    setValue(newValue);
  };
  const handleDownload = (contenido, nombre_archivo) => {
    const element = document.createElement("a");
    const file = new Blob([contenido], { type: "text/plain" });
    element.href = URL.createObjectURL(file);
    element.download = nombre_archivo;
    document.body.appendChild(element);
    element.click();
    document.body.removeChild(element);
  }

  const handlerGetReportes = () => {
    Service.reportes().then((data) => {
      console.log(data);
      const newTabs = data.map((reporte, i) => {
        let names = reporte.path.split("/");
        let name_file = names[names.length-1];
        return {
          label: `Reporte ${i + 1}`,
          content: (
            <>
              <h1>Reporte {reporte.type}</h1>
              <h3>Path: {reporte.path}</h3>
              {reporte.type === "FILE" && <button style={{ marginLeft: 'auto', marginRight: 'auto', fontSize:'24px'}} className="btn colorbtn1" type="button" onClick={() => handleDownload(reporte.file, name_file)}><span class="material-symbols-outlined">download</span>Descargar reporte generado</button>}
              <br />
              <br />
              <Graphviz dot={reporte.dot} options={{zoom: true, width: '300%', height: '100%'}}/>
            </>
          ),
        };
      });
      setTabs(newTabs);
    });
  };
  
  useEffect(() => {
    handlerGetReportes();
  }, []);

  return (
    <Box sx={{ flexGrow: 1, bgcolor: '#91E3FF', display: 'flex', height: '100%', overflow: 'auto' }}>
      <Tabs
        orientation="vertical"
        variant="scrollable"
        value={value}
        onChange={handleChange}
        aria-label="Vertical tabs example"
        sx={{ borderRight: 1, borderColor: 'divider' }}
      >
        {tabs.map((tab, index) => (
          <Tab label={tab.label} {...a11yProps(index)} key={index} />
        ))}
      </Tabs>
      {tabs.map((tab, index) => (
        <TabPanel value={value} index={index} key={index}>
          {tab.content}
        </TabPanel>
      ))}
    </Box>
  );
}
