namespace LitterRobot.Models.Shared
{
    /// <summary>
    /// The shared key info => slug mapping across the application
    /// </summary>
    public class SlugMapping
    {
        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string LRID { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string Slug { get; set; } = string.Empty;

        /// <inheritdoc />
        public override string ToString() => $"LRID: {this.LRID}, Slug: {this.Slug}";
    }
}
