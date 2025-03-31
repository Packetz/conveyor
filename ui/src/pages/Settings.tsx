import { useState } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  TextField,
  Button,
  Grid,
  Switch,
  FormControlLabel,
  Divider,
  Snackbar,
  Alert,
  Paper,
  Tab,
  Tabs
} from '@mui/material';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`settings-tabpanel-${index}`}
      aria-labelledby={`settings-tab-${index}`}
      {...other}
    >
      {value === index && (
        <Box sx={{ p: 3 }}>
          {children}
        </Box>
      )}
    </div>
  );
}

const Settings = () => {
  const [generalSettings, setGeneralSettings] = useState({
    systemName: 'Conveyor',
    dataDirectory: '/app/data',
    pluginsDirectory: '/app/plugins',
    logsEnabled: true,
    logLevel: 'info',
    autoUpdateEnabled: true,
    maxConcurrentJobs: 5
  });

  const [securitySettings, setSecuritySettings] = useState({
    authEnabled: true,
    adminUsername: 'admin',
    adminPassword: '',
    jwtSecret: '**********',
    jwtExpiry: '24h',
    corsEnabled: false,
    corsOrigins: '*'
  });

  const [notificationSettings, setNotificationSettings] = useState({
    emailEnabled: false,
    emailServer: '',
    emailPort: 587,
    emailUsername: '',
    emailPassword: '',
    emailFrom: '',
    slackEnabled: false,
    slackWebhook: ''
  });

  const [tabValue, setTabValue] = useState(0);
  const [snackbar, setSnackbar] = useState({
    open: false,
    message: '',
    severity: 'success' as 'success' | 'error'
  });

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  const handleGeneralSettingsChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target;
    setGeneralSettings({
      ...generalSettings,
      [name]: type === 'checkbox' ? checked : value
    });
  };

  const handleSecuritySettingsChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target;
    setSecuritySettings({
      ...securitySettings,
      [name]: type === 'checkbox' ? checked : value
    });
  };

  const handleNotificationSettingsChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target;
    setNotificationSettings({
      ...notificationSettings,
      [name]: type === 'checkbox' ? checked : value
    });
  };

  const handleSaveSettings = () => {
    // In a real implementation, we would save these settings to the API
    console.log('Saving settings:', {
      general: generalSettings,
      security: securitySettings,
      notification: notificationSettings
    });

    setSnackbar({
      open: true,
      message: 'Settings saved successfully',
      severity: 'success'
    });
  };

  const closeSnackbar = () => {
    setSnackbar({ ...snackbar, open: false });
  };

  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 3 }}>Settings</Typography>

      <Paper sx={{ mb: 3 }}>
        <Tabs
          value={tabValue}
          onChange={handleTabChange}
          variant="fullWidth"
          indicatorColor="primary"
          textColor="primary"
        >
          <Tab label="General" />
          <Tab label="Security" />
          <Tab label="Notifications" />
        </Tabs>
      </Paper>

      <Card>
        <CardContent>
          <TabPanel value={tabValue} index={0}>
            <Typography variant="h6" sx={{ mb: 2 }}>General Settings</Typography>
            <Grid container spacing={3}>
              <Grid item xs={12} md={6}>
                <TextField
                  fullWidth
                  label="System Name"
                  name="systemName"
                  value={generalSettings.systemName}
                  onChange={handleGeneralSettingsChange}
                  margin="normal"
                />
                <TextField
                  fullWidth
                  label="Data Directory"
                  name="dataDirectory"
                  value={generalSettings.dataDirectory}
                  onChange={handleGeneralSettingsChange}
                  margin="normal"
                />
                <TextField
                  fullWidth
                  label="Plugins Directory"
                  name="pluginsDirectory"
                  value={generalSettings.pluginsDirectory}
                  onChange={handleGeneralSettingsChange}
                  margin="normal"
                />
              </Grid>
              <Grid item xs={12} md={6}>
                <TextField
                  fullWidth
                  label="Log Level"
                  name="logLevel"
                  value={generalSettings.logLevel}
                  onChange={handleGeneralSettingsChange}
                  margin="normal"
                  select
                  SelectProps={{ native: true }}
                >
                  <option value="debug">Debug</option>
                  <option value="info">Info</option>
                  <option value="warn">Warning</option>
                  <option value="error">Error</option>
                </TextField>
                <TextField
                  fullWidth
                  label="Max Concurrent Jobs"
                  name="maxConcurrentJobs"
                  type="number"
                  value={generalSettings.maxConcurrentJobs}
                  onChange={handleGeneralSettingsChange}
                  margin="normal"
                />
                <Box sx={{ mt: 2 }}>
                  <FormControlLabel
                    control={
                      <Switch
                        checked={generalSettings.logsEnabled}
                        onChange={handleGeneralSettingsChange}
                        name="logsEnabled"
                      />
                    }
                    label="Enable Logging"
                  />
                </Box>
                <Box>
                  <FormControlLabel
                    control={
                      <Switch
                        checked={generalSettings.autoUpdateEnabled}
                        onChange={handleGeneralSettingsChange}
                        name="autoUpdateEnabled"
                      />
                    }
                    label="Enable Auto Updates"
                  />
                </Box>
              </Grid>
            </Grid>
          </TabPanel>

          <TabPanel value={tabValue} index={1}>
            <Typography variant="h6" sx={{ mb: 2 }}>Security Settings</Typography>
            <Grid container spacing={3}>
              <Grid item xs={12} md={6}>
                <Box sx={{ mb: 2 }}>
                  <FormControlLabel
                    control={
                      <Switch
                        checked={securitySettings.authEnabled}
                        onChange={handleSecuritySettingsChange}
                        name="authEnabled"
                      />
                    }
                    label="Enable Authentication"
                  />
                </Box>
                <TextField
                  fullWidth
                  label="Admin Username"
                  name="adminUsername"
                  value={securitySettings.adminUsername}
                  onChange={handleSecuritySettingsChange}
                  margin="normal"
                  disabled={!securitySettings.authEnabled}
                />
                <TextField
                  fullWidth
                  label="Admin Password"
                  name="adminPassword"
                  type="password"
                  value={securitySettings.adminPassword}
                  onChange={handleSecuritySettingsChange}
                  margin="normal"
                  placeholder="Enter new password to change"
                  disabled={!securitySettings.authEnabled}
                />
              </Grid>
              <Grid item xs={12} md={6}>
                <TextField
                  fullWidth
                  label="JWT Secret"
                  name="jwtSecret"
                  type="password"
                  value={securitySettings.jwtSecret}
                  onChange={handleSecuritySettingsChange}
                  margin="normal"
                  disabled={!securitySettings.authEnabled}
                />
                <TextField
                  fullWidth
                  label="JWT Expiry"
                  name="jwtExpiry"
                  value={securitySettings.jwtExpiry}
                  onChange={handleSecuritySettingsChange}
                  margin="normal"
                  placeholder="e.g. 24h, 7d"
                  disabled={!securitySettings.authEnabled}
                />
                <Box sx={{ mt: 2 }}>
                  <FormControlLabel
                    control={
                      <Switch
                        checked={securitySettings.corsEnabled}
                        onChange={handleSecuritySettingsChange}
                        name="corsEnabled"
                      />
                    }
                    label="Enable CORS"
                  />
                </Box>
                <TextField
                  fullWidth
                  label="CORS Origins"
                  name="corsOrigins"
                  value={securitySettings.corsOrigins}
                  onChange={handleSecuritySettingsChange}
                  margin="normal"
                  disabled={!securitySettings.corsEnabled}
                  placeholder="* or comma-separated URLs"
                />
              </Grid>
            </Grid>
          </TabPanel>

          <TabPanel value={tabValue} index={2}>
            <Typography variant="h6" sx={{ mb: 2 }}>Notification Settings</Typography>

            <Typography variant="subtitle1" sx={{ mb: 2 }}>Email Notifications</Typography>
            <Box sx={{ mb: 2 }}>
              <FormControlLabel
                control={
                  <Switch
                    checked={notificationSettings.emailEnabled}
                    onChange={handleNotificationSettingsChange}
                    name="emailEnabled"
                  />
                }
                label="Enable Email Notifications"
              />
            </Box>
            <Grid container spacing={3} sx={{ mb: 3 }}>
              <Grid item xs={12} md={6}>
                <TextField
                  fullWidth
                  label="SMTP Server"
                  name="emailServer"
                  value={notificationSettings.emailServer}
                  onChange={handleNotificationSettingsChange}
                  margin="normal"
                  disabled={!notificationSettings.emailEnabled}
                />
                <TextField
                  fullWidth
                  label="SMTP Port"
                  name="emailPort"
                  type="number"
                  value={notificationSettings.emailPort}
                  onChange={handleNotificationSettingsChange}
                  margin="normal"
                  disabled={!notificationSettings.emailEnabled}
                />
              </Grid>
              <Grid item xs={12} md={6}>
                <TextField
                  fullWidth
                  label="SMTP Username"
                  name="emailUsername"
                  value={notificationSettings.emailUsername}
                  onChange={handleNotificationSettingsChange}
                  margin="normal"
                  disabled={!notificationSettings.emailEnabled}
                />
                <TextField
                  fullWidth
                  label="SMTP Password"
                  name="emailPassword"
                  type="password"
                  value={notificationSettings.emailPassword}
                  onChange={handleNotificationSettingsChange}
                  margin="normal"
                  disabled={!notificationSettings.emailEnabled}
                />
                <TextField
                  fullWidth
                  label="From Email Address"
                  name="emailFrom"
                  value={notificationSettings.emailFrom}
                  onChange={handleNotificationSettingsChange}
                  margin="normal"
                  disabled={!notificationSettings.emailEnabled}
                />
              </Grid>
            </Grid>

            <Divider sx={{ my: 3 }} />

            <Typography variant="subtitle1" sx={{ mb: 2 }}>Slack Notifications</Typography>
            <Box sx={{ mb: 2 }}>
              <FormControlLabel
                control={
                  <Switch
                    checked={notificationSettings.slackEnabled}
                    onChange={handleNotificationSettingsChange}
                    name="slackEnabled"
                  />
                }
                label="Enable Slack Notifications"
              />
            </Box>
            <TextField
              fullWidth
              label="Slack Webhook URL"
              name="slackWebhook"
              value={notificationSettings.slackWebhook}
              onChange={handleNotificationSettingsChange}
              margin="normal"
              disabled={!notificationSettings.slackEnabled}
            />
          </TabPanel>

          <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: 3 }}>
            <Button variant="contained" color="primary" onClick={handleSaveSettings}>
              Save Settings
            </Button>
          </Box>
        </CardContent>
      </Card>

      {/* Snackbar for notifications */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={6000}
        onClose={closeSnackbar}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert onClose={closeSnackbar} severity={snackbar.severity}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
};

export default Settings; 