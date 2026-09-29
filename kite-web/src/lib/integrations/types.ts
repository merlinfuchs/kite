// A service blocks can talk to. Blocks reference integrations, see
// ../blocks and design/integrations.md.
export interface Integration {
  id: string;
  name: string;
}
