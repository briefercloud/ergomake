import { Profile } from '../../hooks/useProfile'
import Layout from '../components/Layout'

interface Props {
  profile: Profile
}

const Projects = ({ profile }: Props) => {
  return (
    <Layout profile={profile}>
      <h1>Hello world!</h1>
    </Layout>
  )
}

export default Projects
